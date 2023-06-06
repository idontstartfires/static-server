function reloadStyle(filename) {
    console.log(filename);
    const href = `http://static.localhost/style/${filename}`;
    const link = document.querySelector(`link[href^="${href}"]`);
    link.href = `${href}?reload=${new Date().getTime()}`;
}

self.addEventListener("load", () => {
    const eventSource = new EventSource("http://static.localhost/event/style");

    eventSource.onmessage = event => {
        console.log(event);
        reloadStyle(event.data);
    }

    eventSource.onerror = error => {
        console.log(error);
    }
});
