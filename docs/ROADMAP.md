# Roadmap TourneyRank (2026)

> Estado actual: **Fundamentos Completos (MVP en progreso)**
> Clean Architecture · Go + MongoDB · React + Vite

## 🗺️ Visión Global

Convertir el esqueleto actual en una plataforma funcional para organizar torneos de Warzone (y otros juegos) con rankings automatizados.

---

## 📅 Hitos Estimados

### ✅ Fase 0: Arquitectura y User Mgmt (Completado)
*   [x] Estructura Clean Architecture (Domain, Usecase, Infra).
*   [x] Configuración de Docker (Mongo, Redis).
*   [x] Auth Service (JWT, Login/Register).
*   [x] Games Service (CRUD de Juegos y configuración de schemas).
*   [x] MongoDB Repositories conectados.
*   [x] Frontend Scaffolding (Routing, Store, Styles).

### 🚧 Fase 1: Datos y Rankings (Próximo Objetivo)
**Meta:** Que el sistema pueda recibir stats de partidas y mostrar un ranking real.

*   [ ] **Player Service**: Endpoints para crear/listar jugadores vinculados a usuarios.
*   [ ] **Matches & Stats**:
    *   Endpoint para **ingesta de partidas** (POST `/matches`).
    *   Lógica de validación contra el `StatSchema` del juego.
*   [ ] **Ranking Engine**:
    *   Trigger de recálculo al recibir una partida (Sync o Async).
    *   Mejorar `WarzoneCalculator` (consistencia y normalizaciones con datos por partida).
*   [ ] **Frontend Integration**:
    *   Conectar páginas de Login/Register a la API real.
    *   Dashboard mostrando datos reales del usuario.

### 🔭 Fase 2: Torneos y Equipos (Q2 2026)
**Meta:** Organizar la primera competencia administrada.

*   [ ] **Tournament Domain**: Crear torneos, fases, fechas.
*   [ ] **Teams**: Crear equipos, invitar miembros.
*   [ ] **Inscripciones**: Flujo de un usuario inscribiéndose a un torneo.
*   [ ] **Leaderboard en tiempo real**: Usar WebSocket para actualizar la tabla durante el torneo.

### Phase 3: Automatización y Polish (Q3 2026)
*   [ ] Redis Caching de leaderboards.
*   [ ] Integración con Discord (Notificaciones).
*   [ ] Admin Panel completo en Frontend.

---

## 🚀 Próximos pasos inmediatos (Tu Lista de Tareas)

1.  **Frontend Auth**: Conectar el formulario de Login/Register con el backend (`useAuthStore` + API).
2.  **Player Profile**: Que al loguearme vea mi "Player" (crearlo si no existe).
3.  **Primer Match**: Crear un script o endpoint para simular una partida y ver cómo cambia el ranking.

