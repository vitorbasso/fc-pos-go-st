# fc-pos-go-st
Segundo desafio pós go expert (stress test)

## Requerimentos
  * golang versão 1.22.3 ou superior
  * ou docker

## Rodando o projeto
  * Pode rodar com `go run main.go` mais as flags necessárias. (Ou então após buildar o projeto com `go build` com `strest` mais as flags)
  * Ou então, após buildar a imagem com `docker build -t strest .` na raiz do projeto, pode rodar com `docker run strest` mais as flags necessárias

## Funcionamento
  * Ele realiza chamadas para a url especificada na flag `--url` ou `-u` (Obrigatória)
  * Realiza uma quantidade de requests especificada pela flag `--requests` ou `-r` (Default: 1)
  * Realiza de forma concorrente com um número especificado pela flag `--concurrency` ou `-c` (Default: 1)

  Exemplo:
  ```bash
  docker build -t strest .
  docker run strest -u http://localhost:8080 -r 100 -c 10
  ```
  ```bash
  go run main.go -u http://localhost:8080 -r 100 -c 10
  ```
  ```bash
  ./strest -u http://localhost:8080 -r 100 -c 10
  ```
  ### Outras configurações disponíveis
  * Configura-se o método http por `--method` ou `-m` (Default: GET)
  * Configura-se o corpo por `--body` ou `-b` (Default: nil)
  * configura-se headers por `--header` ou `-H` para cada header individualmente
  * Configura-se o timeout das requisições por `--timeout` ou `-t` (Default: 10s)
  * Configura-se a quantidade de requests de warmup por `--warmup` ou `-w` (Default: 0)

## Exemplo de resposta
![image](/docs/images/Screenshot%202024-06-03%20122750.png)


