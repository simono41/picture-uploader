# Picture Uploader

This project is a simple web application written in Go for uploading and viewing images.

## Getting Started

To run this application, follow the steps below:

### Prerequisites

- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/)

### Instructions

1. Clone the repository:

    ```bash
    git clone <repository-url>
    ```

2. Navigate to the project directory:

    ```bash
    cd <project-directory>
    ```

3. Create a folder named `uploads`:

    ```bash
    mkdir uploads
    ```

4. Set permissions for the `uploads` folder:

    ```bash
    chmod 777 uploads
    ```

5. Run the application using Docker Compose:

    ```bash
    docker-compose up -d
    ```

The application should be accessible at [http://localhost:8080](http://localhost:8080).

## Configuration

Modify the `docker-compose.yml` file to adjust environment variables, ports, or any other configurations as needed.

## Upload via Terminal

    ~~~
    curl -X POST -F "image=@/pfad/zur/datei/bild.jpg" http://localhost:8080/upload
    ~~~

Ersetzen Sie /pfad/zur/datei/bild.jpg durch den tatsächlichen Pfad zu Ihrer Datei und http://localhost:8080/upload durch die URL Ihres Servers und den Endpunkt für den Dateiupload.

Hier ist eine Erläuterung der Optionen, die in der Curl-Anfrage verwendet werden:

    ~~~
    -X POST: Legt die HTTP-Methode auf POST fest, was in diesem Fall verwendet wird, um die Datei hochzuladen.
    -F "image=@/pfad/zur/datei/bild.jpg": Teilt Curl mit, dass es sich um ein Formular-Upload handelt (-F), und gibt den Namen des Formularfelds (“image”) sowie den Dateipfad (@/pfad/zur/datei/bild.jpg) an.
    http://localhost:8080/upload: Die URL des Servers und des Endpunkts, an den die Datei hochgeladen werden soll.
    ~~~

Führen Sie diese Curl-Anfrage in einem Terminal aus, und die Datei wird an den angegebenen Server hochgeladen.

## Additional Information

- This project uses NGINX as a reverse proxy. Ensure that the required networks (`nginx-proxy` and `edge`) are set up externally or adjust the `docker-compose.yml` accordingly.
- If you encounter issues with image uploads, verify the permissions on the `uploads` folder.

### Support and Issues

For support or to report issues, please [open an issue](<repository-url>/issues).

### License

This project is licensed under the [MIT License](LICENSE).
