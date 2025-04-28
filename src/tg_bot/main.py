import asyncio
import re
import logging
import os
import json
from typing import Optional, Any, Dict, List

import aiohttp
from aiogram import Bot, Dispatcher, F
from aiogram.filters import CommandStart
from aiogram.types import (
    Message,
    CallbackQuery,
    InlineKeyboardMarkup,
    InlineKeyboardButton,
    InputMediaPhoto,
)
from aiogram.fsm.context import FSMContext
from aiogram.fsm.storage.memory import MemoryStorage
from aiogram.fsm.state import StatesGroup, State
from aiogram.exceptions import TelegramBadRequest

# Адрес бэкенда на порту 9000
API_URL = "http://127.0.0.1:9000/api/v2"
# Токен Telegram-бота
BOT_TOKEN = os.getenv("TG_BOT_TOKEN")
if not BOT_TOKEN:
    raise RuntimeError("Переменная окружения TG_BOT_TOKEN не установлена")

# Настройка логирования
logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s | %(levelname)-8s | %(name)s | %(message)s"
)
logger = logging.getLogger(__name__)

# FSM-состояния
class Form(StatesGroup):
    root_name = State()
    new_list_name = State()
    rename_list = State()
    confirm_delete = State()
    add_poster = State()

class APIError(Exception):
    def __init__(self, status: int, message: str):
        super().__init__(f"{status}:{message}")
        self.status = status
        self.message = message

class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self

    async def __aexit__(self, exc_type, exc, tb):
        if self.session:
            await self.session.close()

    async def _request(self, method: str, path: str, token: str, **kwargs: Any) -> Any:
        url = f"{self.base_url}{path}"
        headers = {"X-User-Token": token}
        logger.debug(f"> {method} {url}")
        logger.debug(f"Headers: {headers}")
        if 'json' in kwargs:
            logger.debug(f"Payload: {kwargs['json']}")
        async with self.session.request(method, url, headers=headers, **kwargs) as resp:
            text = await resp.text()
            logger.debug(f"< {resp.status} {url}")
            logger.debug(f"Resp text: {text}")
            if resp.status >= 400:
                raise APIError(resp.status, text)
            try:
                data = json.loads(text)
                logger.debug(f"Parsed JSON: {data}")
                return data
            except json.JSONDecodeError:
                return text

    async def signup(self, token: str) -> int:
        return await self._request('POST', '/sign-up', token, json={})

    async def get_root_list(self, token: str) -> Dict[str, Any]:
        return await self._request('GET', '/lists-root', token)

    async def get_list(self, token: str, list_id: int) -> Dict[str, Any]:
        return await self._request('GET', f'/lists/{list_id}', token)

    async def create_list(self, token: str, name: str, parent_id: Optional[int]) -> Dict[str, Any]:
        payload = {"name": name}
        if parent_id is not None:
            payload["parentId"] = parent_id
        return await self._request('POST', '/lists', token, json=payload)

    async def update_list(self, token: str, list_id: int, new_name: str) -> Any:
        return await self._request('PUT', f'/lists/{list_id}', token, json={"name": new_name})

    async def delete_list(self, token: str, list_id: int) -> Any:
        return await self._request('DELETE', f'/lists/{list_id}', token)

    async def get_sublists(self, token: str, parent_id: int) -> List[Dict[str, Any]]:
        return await self._request('GET', f'/sublists/{parent_id}', token)

    async def get_list_posters(self, token: str, list_id: int) -> List[Dict[str, Any]]:
        return await self._request('GET', f'/lists/{list_id}/posters', token)

    async def get_poster(self, token: str, poster_id: int) -> Dict[str, Any]:
        return await self._request('GET', f'/posters/{poster_id}', token)

    async def create_poster_kp(self, token: str, kp_id: str) -> Dict[str, Any]:
        return await self._request('POST', '/posters/kp', token, json={"kp_id": kp_id})

    async def add_poster_to_list(self, token: str, list_id: int, poster_id: int) -> Any:
        return await self._request('POST', f'/lists/{list_id}/posters/{poster_id}', token)

    async def delete_poster_from_list(self, token: str, list_id: int, poster_id: int) -> Any:
        return await self._request('DELETE', f'/lists/{list_id}/posters/{poster_id}', token)

    async def create_poster_record(self, token: str, poster_id: int) -> Dict[str, Any]:
        return await self._request('POST', f'/poster-records/{poster_id}', token)

    async def delete_poster_record(self, token: str, poster_id: int) -> Any:
        return await self._request('DELETE', f'/poster-records/{poster_id}', token)

async def main():
    bot = Bot(token=BOT_TOKEN)
    dp = Dispatcher(storage=MemoryStorage())

    @dp.message(CommandStart())
    async def on_start(message: Message, state: FSMContext):
        user_token = str(message.from_user.id)
        async with APIClient(API_URL) as api:
            try:
                await api.signup(user_token)
            except APIError as e:
                if e.status != 409:
                    logger.error(f"Signup error: {e}")
        try:
            async with APIClient(API_URL) as api:
                root = await api.get_root_list(user_token)
        except APIError as e:
            if e.status == 404:
                await message.answer("У вас еще нет корневого списка. Введите его название:")
                await state.set_state(Form.root_name)
                return
            await message.answer("Ошибка при получении списка.")
            return
        await show_list(message, root['id'], user_token)

    @dp.message(Form.root_name)
    async def process_root_name(message: Message, state: FSMContext):
        user_token = str(message.from_user.id)
        name = message.text.strip()
        try:
            async with APIClient(API_URL) as api:
                resp = await api.create_list(user_token, name, None)
        except APIError as e:
            await message.answer(f"Не удалось создать список: {e.message}")
            await state.clear()
            return
        root_id = resp.get('id')
        await message.answer(f"Список '{name}' создан.")
        await state.clear()
        await show_list(message, root_id, user_token)

    async def show_list(event: Message | CallbackQuery, list_id: int, token: str):
        async with APIClient(API_URL) as api:
            info = await api.get_list(token, list_id)
            name = info.get('name', '')
            sublists = await api.get_sublists(token, list_id)
            try:
                posters_raw = await api.get_list_posters(token, list_id)
            except APIError:
                posters_raw = []
            posters: List[Dict[str, Any]] = []
            for p in posters_raw:
                pid = p.get('posterId')
                try:
                    poster_info = await api.get_poster(token, pid)
                    title = poster_info.get('name', 'Unknown')
                except APIError:
                    title = 'Unknown'
                posters.append({'id': pid, 'name': title})
        rows: List[List[InlineKeyboardButton]] = []
        if not sublists and not posters:
            text = f"🎬 Список '{name}' пуст!"
            rows.append([
                InlineKeyboardButton(text="➕ Добавить фильм", callback_data=f"add_poster_{list_id}"),
                InlineKeyboardButton(text="➕ Новый подсписок", callback_data=f"new_sub_{list_id}")
            ])
        else:
            text = f"📂 {name}"
            for sub in sublists:
                rows.append([InlineKeyboardButton(text=f"📁 {sub['name']}", callback_data=f"list_{sub['id']}" )])
            for p in posters:
                rows.append([InlineKeyboardButton(text=f"🎥 {p['name']}", callback_data=f"poster_{list_id}_{p['id']}" )])
            rows.append([InlineKeyboardButton(text="➕ Добавить фильм", callback_data=f"add_poster_{list_id}")])
            rows.append([
                InlineKeyboardButton(text="➕ Новый подсписок", callback_data=f"new_sub_{list_id}"),
                InlineKeyboardButton(text="✏️ Переименовать", callback_data=f"rename_{list_id}"),
                InlineKeyboardButton(text="🗑️ Удалить", callback_data=f"delete_{list_id}")
            ])
        rows.append([
            InlineKeyboardButton(text="⬅️ Назад", callback_data=f"back_{list_id}"),
            InlineKeyboardButton(text="🏠 Главный", callback_data="home")
        ])
        kb = InlineKeyboardMarkup(inline_keyboard=rows)
        if isinstance(event, Message):
            await event.answer(text, reply_markup=kb)
        else:
            try:
                await event.message.edit_text(text, reply_markup=kb)
            except TelegramBadRequest:
                await event.message.answer(text, reply_markup=kb)

    @dp.callback_query(F.data.startswith('back_'))
    async def on_back(callback: CallbackQuery):
        parts = callback.data.split('_')
        current_id = int(parts[-1])
        token = str(callback.from_user.id)
        await callback.answer()
        async with APIClient(API_URL) as api:
            try:
                current = await api.get_list(token, current_id)
                parent_id = current.get('parentId')
                if parent_id:
                    await show_list(callback, parent_id, token)
                else:
                    root = await api.get_root_list(token)
                    await show_list(callback, root['id'], token)
            except APIError:
                root = await api.get_root_list(token)
                await show_list(callback, root['id'], token)

    @dp.callback_query(F.data == 'home')
    async def on_home(callback: CallbackQuery):
        token = str(callback.from_user.id)
        await callback.answer()
        async with APIClient(API_URL) as api:
            root = await api.get_root_list(token)
        await show_list(callback, root['id'], token)

    @dp.callback_query(F.data.startswith('rename_'))
    async def ask_rename(callback: CallbackQuery, state: FSMContext):
        list_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Введите новое название списка:")
        await state.update_data(rename_list_id=list_id)
        await state.set_state(Form.rename_list)

    @dp.message(Form.rename_list)
    async def process_rename(message: Message, state: FSMContext):
        data = await state.get_data()
        list_id = data.get('rename_list_id')
        new_name = message.text.strip()
        token = str(message.from_user.id)
        async with APIClient(API_URL) as api:
            try:
                await api.update_list(token, list_id, new_name)
            except APIError as e:
                await message.answer(f"Ошибка переименования: {e.message}")
                await state.clear()
                return
        await message.answer(f"Список переименован в '{new_name}'.")
        await state.clear()
        await show_list(message, list_id, token)

    @dp.callback_query(F.data.startswith('new_sub_'))
    async def ask_new_sub(callback: CallbackQuery, state: FSMContext):
        parent_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Введите название нового подсписка:")
        await state.update_data(parent_id=parent_id)
        await state.set_state(Form.new_list_name)

    @dp.message(Form.new_list_name)
    async def process_new_sub(message: Message, state: FSMContext):
        data = await state.get_data()
        parent_id = data.get('parent_id')
        user_token = str(message.from_user.id)
        name = message.text.strip()
        async with APIClient(API_URL) as api:
            try:
                resp = await api.create_list(user_token, name, parent_id)
            except APIError as e:
                await message.answer(f"Ошибка создания подсписка: {e.message}")
                await state.clear()
                return
        new_id = resp.get('id')
        await message.answer(f"Подсписок '{name}' создан.")
        await state.clear()
        await show_list(message, new_id, user_token)

    @dp.callback_query(F.data.startswith('poster_'))
    async def show_poster_actions(callback: CallbackQuery):
        parts = callback.data.split('_')
        list_id, poster_id = int(parts[1]), int(parts[2])
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            try:
                info = await api.get_poster(token, poster_id)
            except APIError:
                await callback.answer("Не удалось получить данные постера.")
                return
        # Формируем подпись на русском с эмодзи
        caption = (
            f"🎬 {info['name']} ({info['year']})\n"
            f"⏱️ Хронометраж: {info['chrono']} мин\n"
            f"📅 Создано: {info['createdat']}\n"
            f"🎭 Жанры: {', '.join(info['genres'])}\n"
            f"🔗 KP ID: {info['kp_id']}"
        )
        kb = InlineKeyboardMarkup(inline_keyboard=[
            [
            InlineKeyboardButton(text="❌ Удалить", callback_data=f"del_p_{list_id}_{poster_id}"),
            InlineKeyboardButton(text="✅ Просмотрено", callback_data=f"record_{list_id}_{poster_id}")
            ],
            [
                # Везде, где вы строите клавиатуру:
            InlineKeyboardButton(text="⬅️ Назад", callback_data=f"back_{list_id}"),
            InlineKeyboardButton(text="🏠 Главный", callback_data="home"),
            ],
        ])
        await callback.answer()
        image_url = info.get('image_url')
        if image_url:
            try:
                media = InputMediaPhoto(media=image_url, caption=caption)
                await callback.message.edit_media(media, reply_markup=kb)
            except Exception:
                await callback.message.edit_text(caption, reply_markup=kb)
        else:
            await callback.message.edit_text(caption, reply_markup=kb)


    @dp.callback_query(F.data.startswith('add_poster_'))
    async def ask_add_poster(callback: CallbackQuery, state: FSMContext):
        list_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Мы добавим фильм прямо по ссылке из Кинопоиска (из обычного, не HD 🥲), просто пришли нам ее, либо его айдишник (чиселки в конце)😉 :")
        await state.update_data(list_id=list_id)
        await state.set_state(Form.add_poster)

    @dp.message(Form.add_poster)
    async def process_add_poster(message: Message, state: FSMContext):
        try:
            data = await state.get_data()
            list_id = data.get('list_id')
            user_token = str(message.from_user.id)
            text = message.text.strip()
            match = re.search(r"(\d+)(?:\D*$)", text)
            kp_id = match.group(1) if match else text
            token = str(message.from_user.id)
            async with APIClient(API_URL) as api:
                created = await api.create_poster_kp(token, kp_id)
                poster_id = created.get('id')
                await api.add_poster_to_list(token, list_id, poster_id)
            await message.answer(f"Фильм добавлен.")
            await state.clear()
            await show_list(message, list_id, token)
        except Exception as e:
            message.reply("В этот раз не обошлось без ошибочки. Может, ты отправил ссылку на Кинопоиск HD? Там просто другие ссылки, они не подойдут 🙃")

    @dp.callback_query(F.data.startswith('del_p_'))
    async def process_delete_poster(callback: CallbackQuery):
        parts = callback.data.split('_')
        list_id = int(parts[-2])
        poster_id = int(parts[-1])
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            await api.delete_poster_from_list(token, list_id, poster_id)
        await callback.answer("Удалено.")
        await show_list(callback, list_id, token)

    @dp.callback_query(F.data.startswith('record_'))
    async def process_record(callback: CallbackQuery):
        parts = callback.data.split('_')
        list_id, poster_id = int(parts[1]), int(parts[2])
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            await api.create_poster_record(token, poster_id)
        await callback.answer("Отмечено.")
        await show_list(callback, list_id, token)

    @dp.callback_query(F.data.startswith('list_'))
    async def on_list(callback: CallbackQuery):
        list_id = int(callback.data.split('_')[1])
        token = str(callback.from_user.id)
        await callback.answer()
        await show_list(callback, list_id, token)



    @dp.callback_query(F.data.startswith('back_'))
    async def on_back(callback: CallbackQuery):
        parts = callback.data.split('_')
        current_id = int(parts[-1])
        token = str(callback.from_user.id)
        await callback.answer()
        async with APIClient(API_URL) as api:
            try:
                current = await api.get_list(token, current_id)
                parent_id = current.get('parentId')
                if parent_id:
                    await show_list(callback, parent_id, token)
                else:
                    root = await api.get_root_list(token)
                    await show_list(callback, root['id'], token)
            except APIError:
                root = await api.get_root_list(token)
                await show_list(callback, root['id'], token)

    @dp.callback_query(F.data == 'home')
    async def on_home(callback: CallbackQuery):
        token = str(callback.from_user.id)
        await callback.answer()
        async with APIClient(API_URL) as api:
            root = await api.get_root_list(token)
        await show_list(callback, root['id'], token)

    await dp.start_polling(bot)

if __name__ == '__main__':
    asyncio.run(main())
