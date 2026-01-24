# ⚔️ Argentum Online Go Server - Launcher CLI

Bienvenido al servidor de Argentum Online desarrollado en Go.

Este proyecto incluye una interfaz de línea de comandos (CLI) llamada `launcher` para gestionar el servidor.
---

## 🚀 Guía de Inicio Rápido

### 1. Requisitos Previos
*   **Go** (versión 1.25 o superior recomendada).
*   Tener el repositorio clonado localmente.

### 2. Instalación del Launcher
Para poder ejecutar el comando `launcher` desde cualquier lugar, instálalo en tu sistema:

```bash
cd ao-go-server
go install ./cmd/launcher
```
> **Nota:** Asegúrate de tener tu `$GOPATH/bin` en la variable de entorno PATH de tu sistema.

### 3. Compilación Manual (Alternativa)
Si prefieres no instalarlo globalmente, puedes compilarlo en la carpeta actual:
```bash
go build -o launcher ./cmd/launcher
```

---

## 🛠️ Comandos Principales

### Gestión del Servidor
Controla el estado del proceso:

*   **Iniciar el servidor:**
    ```bash
    ./launcher start --env dev --port 7666
    ```
    *(Añade `&` al final en Linux para ejecutarlo en segundo plano).*

*   **Ver estado y tiempo activo:**
    ```bash
    ./launcher status
    ./launcher uptime
    ```

*   **Detener el servidor:**
    ```bash
    ./launcher stop            # Cierre inmediato
    ./launcher stop --graceful # Cierre limpio (Guardando datos)
    ```

### Gestión del Mundo y Mapas
Modifica el entorno sin reiniciar:

*   **Recargar un mapa específico:**
    ```bash
    ./launcher world reload map_1
    ```

### Moderación y Conexiones
Gestiona a los usuarios conectados:

*   **Listar conexiones activas:**
    ```bash
    ./launcher conn list
    ```
*   **Expulsar (Kick) a un usuario:**
    ```bash
    ./launcher conn kick --account 1234
    ./launcher conn kick --ip 192.168.1.1
    ```

---

## 💻 Compatibilidad
El `servidor` es multiplataforma (windows, linux, mac)

---

