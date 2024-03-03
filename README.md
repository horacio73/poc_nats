# POC NATS

Esta prova de conceito tem como objetivo explorar as capacidades da plataforma NATS no âmbito da persistência, garantia de entrega e time travel das mensagens.  Mais sobre o NATS na página oficial: https://nats.io/

Foram criadas na linguagem Go, programas separados para publicar mensagens e consumi-las do broker, contidos respectivamente nas pastas "pub" e "sub".  O pacote "util" é um pacote de apoio, usado por ambos.  A pasta "test" contém um programa que permite fazer testes em massa no ambiente.

Cada programa tem seu próprio arquivo de configuração default nomeado "/etc/nats.conf" (que não precisa passar via linha de comando), embora aceite que se passe como parâmetro de entrada qualquer outro, da seguinte forma:

`pub -c ./etc/new_conf.txt`  
`sub -c ./etc/outros_parametros.txt`

Na primeira execução do conjunto, para que funcione adequadamente, o programa publicador deve ser executado primeiro.  Ele é o responsável por criar o stream no NATS, caso ainda não exista.  O programa subscritor cria um consumidor durável associado ao referido stream, se ele não já estiver por lá.

Pode-se criar diversos subscritores em paralelo com a mesma identificação (parâmetro "consumer_name" do arq. de conf.) para que se conectem à mesma visão do stream, cada um receberá apenas uma mensagem.  Isto possibilita escalabilidade horizontal no processamento das mensagens (ver seção "Horizontally scalable pull consumers with batching" em https://docs.nats.io/nats-concepts/jetstream).
