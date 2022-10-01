# Synchronizer

## Reviewers

- github.com/cagodoy
- github.com/mtavano
- github.com/NicoBarragan

## Definición

Synchronizer es un servicio diseñado para la sincronización de eventos de un contrato en una base de datos off-chain.

### Contexto

El servicio fue diseñado a partir de las dificultades encontradas al usar el servicio de Moralis. Por otro lado, el propósito de esta solución es que un usuario pueda utilizar su propia nube, pagando los costos reales por utilizar este tipo de servicio a su proveedor de nube.

### ¿Por qué hacemos esto?

El objetivo de dicha solución es debido a que consultar a la cadena de bloques es lento y tiene costos asociados, más aún si es que se realiza desde el frontend, restando UX por tiempos de espera por parte del usuario.

Por otro lado no es flexible en la forma de hacer `queries` ya que solo permite filtrar por lo que teniendo la información en una base de datos añade una capa de simplificación para los usuarios que consuman el servicio.

### Diagramas

Los diagramas de arquitectura se encuentran en el siguiente link: https://app.diagrams.net/?src=about#G1PxvFkkQKAgMXKkp0dnIGnzweYV3uBXMT

### Solución propuesta V1 (DFP)

Es la solución diseñada para el producto `DeFi for People`. Fue implementada usando el lenguaje TypeScript. Si bien funcionó para el cometido, contaba una serie de restricciones debido a que se usó una base de datos relacional, lo que implicó hacer modelos a nivel de código, por ende la solución era específica para el producto `DFP`.

Dicha solución sirvió como base para pensar en una solución abstracta y que pueda ser usada por distintos consumidores independiente al smart contract que estén trabajando.

### Solución propuesta V2 (Cron/DB/API)

Es el primer MVP implementado por `Darch Labs`. La solución implementa en código `Golang` un Cron que, a partir de un ABI ingresado, consigue la información relacionada a los eventos que fueron especificados en dicho ABI.

El Cron será gatillado utilizando el lenguaje de programación, por lo que se deberá tener precauciones en caso de que se rompa el servicio mientras escucha un evento, ya que si el servicio se cae, afecta la escucha de todos los eventos.

Se proveerá una interfaz de uso para que se pueda administrar los eventos, tales como: obtener eventos, agregar eventos y remover eventos.

## Diagramas

## TODO

- [ ] Construir UI para manejo de servicio via app
