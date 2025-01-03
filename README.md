1. Установить go с официального сайта Go https://go.dev/dl/
2. Скачать проект, можно через ```git clone https://github.com/Dorafanboy/reddio_soft.git```
3. Установить зависимости, ```go mod tidy```
4. Заполнить данные в папке data.
Приватные ключи заполнить в evm_private_keys.txt, можно указывать с '0x' или без, 1 строка 1 приватный ключ, также указать приватники в eclipse_private_keys.txt прокси в proxies.txt указывать в формате username:login@ip:port, в data/config.yaml менять конфиг.
6. ```make run``` чтобы запустить скрипт

Количество приватников evm и eclipse должно совпадать. В конфиге можно указать в thread при желании запуск в несколько потоков. Прокси равномерно распределяются между всеми аккаунтами. Распределение рассчитывается по формуле: `аккаунтов_на_прокси = всего_аккаунтов / всего_прокси`

Софт выполняет свапы на:
1. Orca.
2. Lifinity.
3. Invariant.
4. Solar.
5. Создает коллекцию на Underdog.
6. Делает бридж через Relay из L2 из конфига в Eclipse.
