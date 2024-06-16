# gossip-glomers-challenges

Repositório destinado a conter desafios relacionados à plataforma [Gossip Glumers](https://fly.io/dist-sys/).

## Dasafio 1 - Echo

Nesse desafio, a ferramenta [Maelstrom](https://github.com/jepsen-io/maelstrom/tree/v0.2.3) é utilizada para levantar um nó com um binário configurado para receber mensagens e ecoar respostas a essas mensagens.

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/echo/files).

## Desafio 2 - Unique ID Generation

Nesse desafio, um binário configurado para gerar ids únicos foi implementado.

A ferramenta [Maelstrom](https://github.com/jepsen-io/maelstrom/tree/v0.2.3) então criou 3 nós do binário e realizou 1000 chamadas por segundo num período de 30 segundos, a fim de verificar a geração única de ids num ambiente que simula sistemas distribuídos.

O binário foi configurado para geração de ids númericos com a utilização do banco de dados chave-valor [Redis](https://redis.io/).

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/unique-id-generation/results).

## Desafio 3 - Single-Node Broadcast

Nesse desafio, um binário é configurado para receber mensagens e disponibilizá-las para outros nós.

Os logs resultantes do desafio podem ser visto nos arquivos encontrados na pasta [results](https://github.com/SousaGLucas/gossip-glomers-challenges/tree/main/single-node-broadcast/results).
