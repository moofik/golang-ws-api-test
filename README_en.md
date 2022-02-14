1. Run next command at the root of the project to run MySQL container: docker/up 
2. Run next command at the root of the project to run Golang app: sh run.sh
3. Run next command at the root of the project to init database: docker/init 
4. Websocket Api URL example: http://localhost:8080/ws?fsyms=ETH&tsyms=USD,EUR
5. REST Api URL example: http://localhost:8080/api?fsyms=BTC&tsyms=USD,EUR
6. There are URLs which you can use to test websocket API:
    1. http://localhost:8080/client1 - loads page which connects to WS API with next params: ?fsyms=BTC&tsyms=USD,EUR
    2. http://localhost:8080/client2 - loads page which connects to WS API with next params: ?fsyms=ETH&tsyms=USD,EUR