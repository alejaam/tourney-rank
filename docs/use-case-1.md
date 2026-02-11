A continuacion se detallan las funcionalidades primpordiales para este MVP.

- Usuario de tiplo "player" puede:
    - Registrarse
    - Iniciar sesión
    - Ver su perfil
    - Ver torneos disponibles
    - Inscribirse a un torneo
    - Reportar resultados de partidas
- Usuario de tipo "admin" puede:
    - Iniciar sesión
    - Crear torneos
    - Ver torneos creados
    - Editar torneos creados
    - Eliminar torneos creados
    - Ver reportes de partidas
    - Aceptar o rechazar player en el torneo

El flujo y descripcion de un torneo de warzone en partida privada es el siguiente:
1. El admin crea un torneo con una fecha de inicio y fin, una descripcion y un cupo maximo de jugadores.
2. Los jugadores pueden ver el torneo y decidir inscribirse.
   1. Precondiciones para inscribirse a un torneo:
      1. El jugador debe tener una cuenta y haber iniciado sesión.
      2. El torneo debe estar en status "open".
      3. El torneo no debe haber alcanzado su cupo maximo de jugadores.
      4. El torneo debe ser compatible con el juego y la plataforma del jugador.
      5. El jugador debe cumplir con los requisitos de edad y region establecidos por el torneo.
      6. El jugador no debe estar inscrito en otro torneo que se solape en fechas con el torneo al que quiere inscribirse.
      7. El capitan del equipo debe ser el encargado de inscribir a todo el equipo, y todos los miembros del equipo deben cumplir con las precondiciones anteriores.
   2. El torneo puede estar en status "open" o "closed". Solo los torneos en status "open" pueden recibir inscripciones.
   3. de closed solo espera a iniciar el torneo, una vez iniciado el torneo pasa a status "in progress" y no puede recibir mas inscripciones.
   4. El torneo puede ser duos, trios o squads, dependiendo del juego y la cantidad de jugadores que se quieran inscribir.
   5. En un inicio estara basado en warzone.
   6. Los tipos de torneo son: kill race, privadas battle royale, matchpoint, killpoint, switcharoo, etc.
3. El admin acepta o rechaza a los jugadores inscritos.
4. El admin puede publicar codigo de sala para que los jugadores puedan unirse a la partida privada del torneo.
5. Los jugadores juegan la partida privada del torneo y reportan los resultados, hasta cumplir las N partidas requeridas para el torneo.
6. El sistema calcula las posiciones de los jugadores sacando informacion de kills, placement, etc. y publica las posiciones en el ranking del torneo.