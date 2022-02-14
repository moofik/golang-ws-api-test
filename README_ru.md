1. Запустите следующую комманду из корня проекта для запуска MySQL контейнера: docker/up 
2. Запустите следующую комманду из корня проекта для запуска основного приложения Golang: sh run.sh
3. Запустите следующую комманду из корня проекта для инициализации базы данных: docker/init 
4. Пример URL для Websocket API: http://localhost:8080/ws?fsyms=ETH&tsyms=USD,EUR
5. Пример URL для REST API: http://localhost:8080/api?fsyms=BTC&tsyms=USD,EUR
6. Следующие URL-ы могут использоваться для тестирования Websocket API:
    1. http://localhost:8080/client1 - загружает страницу, которая подключается к websocket api со следующими параметрами: ?fsyms=BTC&tsyms=USD,EUR
    2. http://localhost:8080/client2 - загружает страницу, которая подключается к websocket api со следующими параметрами: ?fsyms=ETH&tsyms=USD,EUR