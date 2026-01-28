# ⚔️ Argentum Online Go Server ⚔️

Implementación moderna y de alto rendimiento del servidor de **Argentum Online** escrita en **Go**. Este proyecto busca renovar la arquitectura del juego clásico, ofreciendo mayor estabilidad, concurrencia nativa, facilidad de despliegue y herramientas de gestión avanzadas para administradores y desarrolladores.

---

## 📋 Requisitos

*   **Go**: Versión 1.25.6 o superior.
*   **Git**: Para clonar el repositorio.
*   **Sistema Operativo**: Compatible con Linux, Windows y macOS.

## 🛠️ Instalación

1.  **Clonar el repositorio:**
    ```bash
    git clone https://github.com/tu-usuario/ao-go-server.git
    cd ao-go-server
    ```

2.  **Descargar dependencias:**
    ```bash
    go mod download
    ```

3.  **Compilar el Launcher:**
    Para facilitar la gestión, compilamos la herramienta CLI principal:
    ```bash
    go build -o aog_launcher ./cmd/aog_launcher
    ```

---

## ✨ Features Principales

*   **aog_launcher**: CLI potente y unificada para el inicio, parada, reinicio y gestión integral del ciclo de vida del servidor.
*   **TUI (Text User Interface)**: Panel de control visual en terminal para monitorear el estado del servidor, logs y métricas en tiempo real sin salir de la consola.
*   **API de Administración**: Interfaz programable para integrar herramientas externas o paneles web.
*   **Monitor de Red**: Observabilidad detallada del tráfico de red, paquetes y conexiones activas.
*   **Hot Reloading**: Capacidad de recargar mapas (`.map`, `.dat`) y archivos de configuración (balances, `server.yaml`) sin detener el servidor.
*   **Inicio Rápido y Caching**: Carga optimizada de recursos y mapas utilizando concurrencia y sistemas de caché para reducir drásticamente los tiempos de arranque.
*   **Arquitectura Concurrente**: Aprovecha las Goroutines de Go para manejar miles de conexiones y tareas simultáneas de manera eficiente.
*   **Multiplataforma**: Ejecutable nativo en cualquier sistema operativo soportado por Go.

---

## ⚙️ Configuración

La configuración del servidor se gestiona principalmente a través de archivos YAML ubicados en `resources/config_yaml/`.

*   **`server.yaml`**: Configuración general (puerto, base de datos, límites).
*   **`project.yaml`**: Rutas de recursos y configuraciones específicas del proyecto.
*   **`balances.yaml`**: Balance de clases, npcs y objetos.
*   **`maps.yaml`**: Propiedades de los mapas.

Puedes editar estos archivos manualmente o utilizar el launcher para ciertas tareas de gestión.

---

## 🚀 Uso

### Usando `aog_launcher` (Recomendado)

El launcher es la herramienta central para interactuar con el servidor.

*   **Iniciar el servidor:**
    ```bash
    ./aog_launcher start
    ```
    Opcionalmente puedes especificar puerto y entorno:
    ```bash
    ./aog_launcher start --port 7666 --env prod
    ```

*   **Ver estado:**
    ```bash
    ./aog_launcher status
    ./aog_launcher uptime
    ```

*   **Gestión de conexiones:**
    ```bash
    ./aog_launcher conn list
    ./aog_launcher conn kick --account <id>
    ```

*   **Recargar mundo (Hot Reload):**
    ```bash
    ./aog_launcher world reload map_1
    ```

*   **Monitor TUI: (Recomendado)**
    ```bash
    ./aog_launcher monitor
    ```

### Ejecución Directa (Desarrollo)

Si prefieres ejecutar el servidor directamente sin el launcher compilado:

```bash
go run cmd/server/main.go
```

---

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Sigue estos pasos para colaborar:

1.  Haz un **Fork** del repositorio.
2.  Crea una rama para tu feature o fix: `git checkout -b feature/nueva-funcionalidad`.
3.  Realiza tus cambios con **Commits claros** y descriptivos.
4.  Asegúrate de que el código compile.
5.  Abre un **Pull Request (PR)** hacia la rama `main` describiendo técnicamente tus cambios.

---

## 📄 Licencia

Este proyecto está bajo la Licencia **GNU General Public License v3.0**. Consulta el archivo [LICENSE](LICENSE) para más detalles.
