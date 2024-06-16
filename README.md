# gossip-glomers-challenges

Repositório destinado a conter desafios relacionados à plataforma [Gossip Glumers](https://fly.io/dist-sys/).

## Dasafio 1 - [Echo](https://fly.io/dist-sys/1/)

Nesse desafio, a ferramenta [Maelstrom](https://github.com/jepsen-io/maelstrom/tree/v0.2.3) é utilizada para levantar um nó com um binário configurado para receber mensagens e ecoar respostas a essas mensagens.

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/echo/files).

## Desafio 2 - [Unique ID Generation](https://fly.io/dist-sys/2/)

Nesse desafio, um binário configurado para gerar ids únicos foi implementado.

A ferramenta [Maelstrom](https://github.com/jepsen-io/maelstrom/tree/v0.2.3) então criou 3 nós do binário e realizou 1000 chamadas por segundo num período de 30 segundos, a fim de verificar a geração única de ids num ambiente que simula sistemas distribuídos.

O binário foi configurado para geração de ids númericos com a utilização do banco de dados chave-valor [Redis](https://redis.io/).

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/unique-id-generation/results).

## Desafio 3 - [Single-Node Broadcast](https://fly.io/dist-sys/3a/)

Nesse desafio, um binário é configurado para receber mensagens e disponibilizá-las para outros nós.

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/single-node-broadcast/results).

## Desafio 4 - [Grow-Only Counter](https://fly.io/dist-sys/4/)

Nesse desafio, um binário é configurado para receber solicitações para incrementar um contador. Esse contador foi configurado para ser [sequencialmente consistente](https://jepsen.io/consistency/models/sequential).

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/grow-only-counter/results).

## Desafio 5 - [Single-Node Kafka-Style Log](https://fly.io/dist-sys/5a/)

Nesse desafio, um bonário foi configurado para receber e enfileirar mensagens em filas, similar ao que a ferramenta [Kafka](https://kafka.apache.org/) faz.

Foi configurado handlers para leitura, commit e listagem dos últimos ofssets para uma lista de filas.

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/single-node-kafka-style-log/results).
