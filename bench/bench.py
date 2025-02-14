import json
import numpy as np
import matplotlib.pyplot as plt

# 📌 1. Читаем JSON-данные из файла
with open("result.txt", "r") as file:
    raw_data = json.load(file)  # Загружаем JSON

# 📌 2. Преобразуем данные: убираем строки, разделяем по пробелам и приводим к числам
arrays = [[] for _ in range(12)]  # 12 массивов для разных метрик

for group in raw_data:  # Каждый вложенный список (по 10 штук)
    for i, line in enumerate(group):  # Внутри каждого списка строки с числами
        numbers = list(map(int, line.split()))  # Разделяем строку по пробелам и конвертируем в int
        for j, num in enumerate(numbers):  # Распределяем по массивам
            arrays[i * 3 + j].append(num)

# 📌 3. Определяем размерность (X-ось)
max_length = max(len(arr) for arr in arrays)
size = list(range(1, max_length + 1))

# 📌 4. Функция для построения графиков
def plot_graph(y1, y2, ylabel, title, filename, labels):
    plt.figure(figsize=(10, 6))
    plt.ylabel(ylabel)
    plt.xlabel("Номер бенчмарка")
    plt.grid(True)

    # Проверяем, что массивы одинаковой длины
    min_length = min(len(size), len(y1), len(y2))
    x_vals, y1_vals, y2_vals = size[:min_length], y1[:min_length], y2[:min_length]

    plt.plot(x_vals, y1_vals, color="darkmagenta", label=labels[0], marker="^")
    plt.plot(x_vals, y2_vals, color="gold", label=labels[1], marker="*")

    plt.legend(labels)
    plt.title(title)
    plt.savefig(filename)
    plt.show()

# 📌 5. Построение графиков
plot_graph(arrays[0], arrays[3], "NsPerOp", "Add client NsPerOp", "addClientNsPerOp.svg", ["gorm", "sqlx"])
plot_graph(arrays[1], arrays[4], "AllocsPerOp", "Add client AllocsPerOp", "addClientAllocsPerOp.svg", ["gorm", "sqlx"])
plot_graph(arrays[2], arrays[5], "AllocedBytesPerOp", "Add client AllocedBytesPerOp", "addClientAllocedBytesPerOp.svg", ["gorm", "sqlx"])

plot_graph(arrays[6], arrays[9], "NsPerOp", "Get client NsPerOp", "getClientNsPerOp.svg", ["gorm", "sqlx"])
plot_graph(arrays[7], arrays[10], "AllocsPerOp", "Get client AllocsPerOp", "getClientAllocsPerOp.svg", ["gorm", "sqlx"])
plot_graph(arrays[8], arrays[11], "AllocedBytesPerOp", "Get client AllocedBytesPerOp", "getClientAllocedBytesPerOp.svg", ["gorm", "sqlx"])
