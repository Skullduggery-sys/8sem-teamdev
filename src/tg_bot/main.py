import asyncio
import re
import logging
import os
import json
from datetime import datetime
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
from aiogram.exceptions import TelegramBadRequest
from aiogram.fsm.context import FSMContext
from aiogram.fsm.storage.memory import MemoryStorage
from aiogram.fsm.state import StatesGroup, State

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
        if 'json' in kwargs:
            logger.debug(f"Payload: {kwargs['json']}")
        async with self.session.request(method, url, headers=headers, **kwargs) as resp:
            text = await resp.text()
            logger.debug(f"< {resp.status} {url}")
            if resp.status >= 400:
                raise APIError(resp.status, text)
            try:
                return json.loads(text)
            except json.JSONDecodeError:
                return text

    # API методы
    async def signup(self, token: str) -> int:
        return await self._request('POST', '/sign-up', token, json={})
    async def get_root_list(self, token: str) -> Dict[str, Any]:
        return await self._request('GET', '/lists-root', token)
    async def get_list(self, token: str, list_id: int) -> Dict[str, Any]:
        return await self._request('GET', f'/lists/{list_id}', token)
    async def create_list(self, token: str, name: str, parent_id: Optional[int]) -> Dict[str, Any]:
        payload = {'name': name}
        if parent_id is not None:
            payload['parentId'] = parent_id
        return await self._request('POST', '/lists', token, json=payload)
    async def update_list(self, token: str, list_id: int, new_name: str) -> Any:
        return await self._request('PUT', f'/lists/{list_id}', token, json={'name': new_name})
    async def delete_list(self, token: str, list_id: int) -> Any:
        return await self._request('DELETE', f'/lists/{list_id}', token)
    async def get_sublists(self, token: str, parent_id: int) -> List[Dict[str, Any]]:
        return await self._request('GET', f'/sublists/{parent_id}', token)
    async def get_list_posters(self, token: str, list_id: int) -> List[Dict[str, Any]]:
        return await self._request('GET', f'/lists/{list_id}/posters', token)
    async def get_poster(self, token: str, poster_id: int) -> Dict[str, Any]:
        return await self._request('GET', f'/posters/{poster_id}', token)
    async def create_poster_kp(self, token: str, kp_id: str) -> Dict[str, Any]:
        return await self._request('POST', '/posters/kp', token, json={'kp_id': kp_id})
    async def add_poster_to_list(self, token: str, list_id: int, poster_id: int) -> Any:
        return await self._request('POST', f'/lists/{list_id}/posters/{poster_id}', token)
    async def delete_poster_from_list(self, token: str, list_id: int, poster_id: int) -> Any:
        return await self._request('DELETE', f'/lists/{list_id}/posters/{poster_id}', token)
    async def create_poster_record(self, token: str, poster_id: int) -> Dict[str, Any]:
        return await self._request('POST', f'/poster-records/{poster_id}', token)
    async def list_poster_records(self, token: str) -> List[Dict[str, Any]]:
        return await self._request('GET', '/poster-records', token)
    async def delete_poster_record(self, token: str, poster_id: int) -> Any:
        return await self._request('DELETE', f'/poster-records/{poster_id}', token)

async def main():
    bot = Bot(token=BOT_TOKEN)
    dp = Dispatcher(storage=MemoryStorage())

    @dp.message(CommandStart())
    async def on_start(message: Message, state: FSMContext):
        user_token = str(message.from_user.id)
        new_user = False
        # пытаемся зарегать юзера
        async with APIClient(API_URL) as api:
            try:
                await api.signup(user_token)
                new_user = True
            except APIError:
                pass

        if new_user:
            await message.answer(
                "👋 Привет, киноман! Добро пожаловать в FilmOlistBot! 🍿\n"
                "Здесь ты создаёшь списки фильмов, добавляешь кино и отмечаешь просмотренное. 🎬\n"
                "Поехали — твой главный список уже готов! 💡"
            )
            async with APIClient(API_URL) as api:
                resp = await api.create_list(user_token, "Главный список", None)
                root_id = resp['id']
            await show_list(message, root_id, user_token)
            return

        try:
            async with APIClient(API_URL) as api:
                root = await api.get_root_list(user_token)
        except APIError as e:
            await message.answer("Ошибка при получении списка.")
            return

        await show_list(message, root['id'], user_token)

    @dp.message(Form.root_name)
    async def process_root_name(message: Message, state: FSMContext):
        token = str(message.from_user.id)
        name = message.text.strip()
        async with APIClient(API_URL) as api:
            resp = await api.create_list(token, name, None)
        await state.clear()
        await show_list(message, resp['id'], token)

    async def show_list(event: Message | CallbackQuery, list_id: int, token: str):
        async with APIClient(API_URL) as api:
            info = await api.get_list(token, list_id)
            name = info.get('name', '')
            parent_id = info.get('parentId')
            root = await api.get_root_list(token)
            root_id = root['id']
            sublists = await api.get_sublists(token, list_id)
            try:
                posters_raw = await api.get_list_posters(token, list_id)
            except APIError:
                posters_raw = []
            posters: List[Dict[str, Any]] = []
            for p in posters_raw:
                pid = p['posterId']
                try:
                    poster = await api.get_poster(token, pid)
                except APIError:
                    poster = {}
                posters.append({
                    'id': pid,
                    'name': poster.get('name', 'Unknown')
                })
        # Формируем клавиатуру
        rows: List[List[InlineKeyboardButton]] = []
        if not sublists and not posters:
            text = f"🎬 Список '{name}' пуст!"
            rows.append([
                InlineKeyboardButton(text="➕ Добавить фильм", callback_data=f"add_poster_{list_id}"),
                InlineKeyboardButton(text="➕ Новый подсписок", callback_data=f"new_sub_{list_id}")
            ])
            if list_id != root_id:
                rows.append([
                    InlineKeyboardButton(text="✏️ Переименовать", callback_data=f"rename_{list_id}"),
                    InlineKeyboardButton(text="🗑️ Удалить", callback_data=f"delete_{list_id}")
                ])
            else:
                rows.append([
                    InlineKeyboardButton(text="✏️ Переименовать", callback_data=f"rename_{list_id}"),
                ])
        else:
            text = f"📂 {name}"
            for sub in sublists:
                rows.append([InlineKeyboardButton(text=f"📁 {sub['name']}", callback_data=f"list_{sub['id']}" )])
            for p in posters:
                rows.append([InlineKeyboardButton(text=f"🎥 {p['name']}", callback_data=f"poster_{list_id}_{p['id']}" )])
            rows.append([InlineKeyboardButton(text="➕ Добавить фильм", callback_data=f"add_poster_{list_id}")])
            if list_id != root_id:
                rows.append([
                    InlineKeyboardButton(text="➕ Новый подсписок", callback_data=f"new_sub_{list_id}"),
                    InlineKeyboardButton(text="✏️ Переименовать", callback_data=f"rename_{list_id}"),
                    InlineKeyboardButton(text="🗑️ Удалить", callback_data=f"delete_{list_id}")
                ])
            else:
                rows.append([
                    InlineKeyboardButton(text="➕ Новый подсписок", callback_data=f"new_sub_{list_id}"),
                    InlineKeyboardButton(text="✏️ Переименовать", callback_data=f"rename_{list_id}")
                ])
        # Навигация и история
        nav_row: List[InlineKeyboardButton] = []
        if list_id != root_id:
            nav_row.append(
                InlineKeyboardButton(text="⬅️ Назад", callback_data=f"back_{list_id}")
            )
            nav_row.append(
                InlineKeyboardButton(text="🏠 Главный", callback_data="home")
            )
        # История оставляем в любом случае
        rows.append([
            InlineKeyboardButton(text="📖 История", callback_data="history")
        ])
        rows.append(nav_row)
        kb = InlineKeyboardMarkup(inline_keyboard=rows)
        if isinstance(event, Message):
            await event.answer(text, reply_markup=kb)
        else:
            try:
                await event.message.edit_text(text, reply_markup=kb)
            except TelegramBadRequest:
                await event.message.answer(text, reply_markup=kb)

    # Создание подсписка
    @dp.callback_query(F.data.startswith('new_sub_'))
    async def ask_new_sub(callback: CallbackQuery, state: FSMContext):
        parent_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Введите название нового списка:")
        await state.update_data(parent_id=parent_id)
        await state.set_state(Form.new_list_name)

    @dp.message(Form.new_list_name)
    async def process_new_sub(message: Message, state: FSMContext):
        data = await state.get_data()
        parent_id = data.get('parent_id')
        name = message.text.strip()
        token = str(message.from_user.id)
        async with APIClient(API_URL) as api:
            resp = await api.create_list(token, name, parent_id)
        await state.clear()
        await show_list(message, resp['id'], token)

    # Переименование списка
    @dp.callback_query(F.data.startswith('rename_'))
    async def ask_rename(callback: CallbackQuery, state: FSMContext):
        list_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Введите новое название списка:")
        await state.update_data(list_id=list_id)
        await state.set_state(Form.rename_list)

    @dp.message(Form.rename_list)
    async def process_rename(message: Message, state: FSMContext):
        data = await state.get_data()
        list_id = data.get('list_id')
        new_name = message.text.strip()
        token = str(message.from_user.id)
        async with APIClient(API_URL) as api:
            await api.update_list(token, list_id, new_name)
        await state.clear()
        await show_list(message, list_id, token)

    # Удаление списка
    @dp.callback_query(F.data.startswith('delete_'))
    async def ask_delete(callback: CallbackQuery, state: FSMContext):
        list_id = int(callback.data.split('_')[-1])
        await callback.answer()
        kb = InlineKeyboardMarkup(inline_keyboard=[
            [InlineKeyboardButton(text="✅ Да", callback_data=f"confirm_yes_{list_id}"),
             InlineKeyboardButton(text="❌ Нет", callback_data=f"confirm_no_{list_id}")]
        ])
        await callback.message.edit_text("Вы уверены, что хотите удалить этот список?", reply_markup=kb)
        await state.update_data(list_id=list_id)
        await state.set_state(Form.confirm_delete)

    @dp.callback_query(F.data.startswith('confirm_yes_'))
    async def process_confirm_yes(callback: CallbackQuery, state: FSMContext):
        data = await state.get_data()
        list_id = data.get('list_id')
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            current = await api.get_list(token, list_id)
            parent_id = current.get('parentId')
            await api.delete_list(token, list_id)
            await callback.answer("Список удален.")
            if parent_id:
                await show_list(callback, parent_id, token)
            else:
                root = await api.get_root_list(token)
                await show_list(callback, root['id'], token)
        await state.clear()

    @dp.callback_query(F.data.startswith('confirm_no_'))
    async def process_confirm_no(callback: CallbackQuery, state: FSMContext):
        data = await state.get_data()
        list_id = data.get('list_id')
        token = str(callback.from_user.id)
        await callback.answer("Удаление отменено.")
        await state.clear()
        await show_list(callback, list_id, token)

    # Добавление постера через KP ID или URL
    @dp.callback_query(F.data.startswith('add_poster_'))
    async def ask_add_poster(callback: CallbackQuery, state: FSMContext):
        list_id = int(callback.data.split('_')[-1])
        await callback.answer()
        await callback.message.answer("Мы добавим фильм сразу по ID или ссылке на страницу Кинопоиска (например, 404900 или https://www.kinopoisk.ru/film/404900/):")
        await state.update_data(list_id=list_id)
        await state.set_state(Form.add_poster)

    @dp.message(Form.add_poster)
    async def process_add_poster(message: Message, state: FSMContext):
        data = await state.get_data()
        list_id = data.get('list_id')
        text = message.text.strip()

        match_id = re.fullmatch(r'\d+', text)

        match_url = re.fullmatch(
            r'https?://(?:www\.)?kinopoisk\.ru/film/(\d+)/?$', text
        )
        if match_id:
            kp_id = match_id.group(0)
        elif match_url:
            kp_id = match_url.group(1)
        else:
            await message.answer(
                "❌ Не удалось распознать ID фильма или ссылку на Кинопоиск.\n"
                "Пожалуйста, введите числовой ID (например, 404900) или полную ссылку "
                "https://www.kinopoisk.ru/film/404900/"
            )
            return 

        user_token = str(message.from_user.id)
        try:
            async with APIClient(API_URL) as api:
                created = await api.create_poster_kp(user_token, kp_id)
                poster_id = created.get('id')
                await api.add_poster_to_list(user_token, list_id, poster_id)
        except APIError as e:
            await message.answer(f"Ошибка при добавлении фильма: {e.message}")
            await state.clear()
            return

        await message.answer("🎉 Фильм успешно добавлен")
        await state.clear()
        await show_list(message, list_id, user_token)

    # Экран постера
    @dp.callback_query(F.data.startswith('poster_'))
    async def show_poster_actions(callback: CallbackQuery):
        _, list_str, poster_str = callback.data.split('_')
        list_id, poster_id = int(list_str), int(poster_str)
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            try:
                info = await api.get_poster(token, poster_id)
            except APIError:
                await callback.answer("Не удалось загрузить данные постера.")
                return
        # Формируем подпись
        created = info.get('createdat', '')
        try:
            date = datetime.fromisoformat(created)
            created_fmt = date.strftime('%d.%m.%Y')
        except Exception:
            created_fmt = created
        caption = (
            f"🎬 {info['name']} ({info['year']})\n"
            f"⏱️ Хронометраж: {info.get('chrono', '?')} мин\n"
            f"📅 Добавлено: {created_fmt}\n"
            f"🎭 Жанры: {', '.join(info.get('genres', []))}\n"
            f"🔗 https://www.kinopoisk.ru/film/{info.get('kp_id')}/"
        )
        rows = [
            [InlineKeyboardButton(text="❌ Удалить", callback_data=f"del_p_{list_id}_{poster_id}"),
             InlineKeyboardButton(text="✅ Просмотрено", callback_data=f"record_{list_id}_{poster_id}")],
            [InlineKeyboardButton(text="⬅️ Назад", callback_data=f"list_{list_id}"),
             InlineKeyboardButton(text="🏠 Главное", callback_data="home")]
        ]
        kb = InlineKeyboardMarkup(inline_keyboard=rows)
        await callback.answer()
        image_url = info.get('image_url')
        try:
            if image_url:
                media = InputMediaPhoto(media=image_url, caption=caption)
                await callback.message.edit_media(media=media, reply_markup=kb)
            else:
                raise Exception
        except (TelegramBadRequest, Exception):
            try:
                await callback.message.edit_text(caption, reply_markup=kb)
            except TelegramBadRequest:
                await callback.message.answer(caption, reply_markup=kb)

    # Удаление постера из списка
    @dp.callback_query(F.data.startswith('del_p_'))
    async def process_delete_poster(callback: CallbackQuery):
        parts = callback.data.split('_')
        list_id = int(parts[-2])
        poster_id = int(parts[-1])
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            await api.delete_poster_from_list(token, list_id, poster_id)
        await callback.answer("Фильм удалён.")
        await show_list(callback, list_id, token)

    # Отметить просмотрено
    @dp.callback_query(F.data.startswith('record_'))
    async def process_record(callback: CallbackQuery):
        _, list_id_str, poster_id_str = callback.data.split('_', 2)
        list_id, poster_id = int(list_id_str), int(poster_id_str)
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            await api.create_poster_record(token, poster_id)
            await api.delete_poster_from_list(token, list_id, poster_id)
        await callback.answer("✅ Фильм перемещен в историю просмотров")
        await show_list(callback, list_id, token)

    # История просмотра
    @dp.callback_query(F.data == 'history')
    async def show_history(callback: CallbackQuery):
        token = str(callback.from_user.id)
        try:
            async with APIClient(API_URL) as api:
                records = await api.list_poster_records(token)
            rows: List[List[InlineKeyboardButton]] = []
            text = "📖 История просмотра:" if records else "📖 История пуста."
            for r in records:
                pid = r.get('posterId')
                async with APIClient(API_URL) as api:
                    try:
                        info = await api.get_poster(token, pid)
                        name = info.get('name', 'Unknown')
                    except APIError:
                        name = 'Unknown'
                rows.append([
                    InlineKeyboardButton(text=f"🎥 {name}", callback_data=f"hist_{pid}")
                ])
            rows.append([
                InlineKeyboardButton(text="⬅️ Назад", callback_data="home")
            ])
            await callback.answer()
            kb = InlineKeyboardMarkup(inline_keyboard=rows)
            try:
                await callback.message.edit_text(text, reply_markup=kb)
            except TelegramBadRequest:
                await callback.message.answer(text, reply_markup=kb)
        except APIError:
            await callback.answer("История пока что пуста. Давай исправим?)")

    @dp.callback_query(F.data.startswith('hist_'))
    async def show_history_item(callback: CallbackQuery):
        poster_id = int(callback.data.split('_')[1])
        token = str(callback.from_user.id)
        async with APIClient(API_URL) as api:
            try:
                info = await api.get_poster(token, poster_id)
            except APIError:
                await callback.answer("Не удалось загрузить данные постера.")
                return
        created = info.get('createdat', '')
        try:
            date = datetime.fromisoformat(created)
            created_fmt = date.strftime('%d.%m.%Y')
        except Exception:
            created_fmt = created
        caption = (
            f"🎬 {info['name']} ({info['year']})\n"
            f"⏱️ Хронометраж: {info.get('chrono', '?')} мин\n"
            f"📅 Добавлено: {created_fmt}\n"
            f"🎭 Жанры: {', '.join(info.get('genres', []))}\n"
            f"🔗 https://www.kinopoisk.ru/film/{info.get('kp_id')}/"
        )
        rows = [
            [InlineKeyboardButton(text="⬅️ Назад", callback_data="history"),
             InlineKeyboardButton(text="🏠 Главное", callback_data="home")]
        ]
        kb = InlineKeyboardMarkup(inline_keyboard=rows)
        await callback.answer()
        image_url = info.get('image_url')
        try:
            if image_url:
                media = InputMediaPhoto(media=image_url, caption=caption)
                await callback.message.edit_media(media=media, reply_markup=kb)
            else:
                raise Exception
        except (TelegramBadRequest, Exception):
            try:
                await callback.message.edit_text(caption, reply_markup=kb)
            except TelegramBadRequest:
                await callback.message.answer(caption, reply_markup=kb)

    # Переход по спискам
    @dp.callback_query(F.data.startswith('list_'))
    async def on_list(callback: CallbackQuery):
        list_id = int(callback.data.split('_')[1])
        token = str(callback.from_user.id)
        await callback.answer()
        await show_list(callback, list_id, token)

    # Назад и Главное
    @dp.callback_query(F.data.startswith('back_'))
    async def on_back(callback: CallbackQuery):
        current_id = int(callback.data.split('_')[1])
        token = str(callback.from_user.id)
        await callback.answer()
        async with APIClient(API_URL) as api:
            current = await api.get_list(token, current_id)
            parent_id = current.get('parentId')
            if parent_id:
                await show_list(callback, parent_id, token)
            else:
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
