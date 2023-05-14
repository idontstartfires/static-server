function reloadStyle(filename) {
    let href = `http://static.localhost/style/${filename}`;
    let link = document.querySelector(`link[href^="${href}"]`);
    link.href = `${href}?reload=${new Date().getTime()}`;
}

self.addEventListener("load", () => {
    const eventSource = new EventSource("http://static.localhost/event/style");

    eventSource.onmessage = event => {
        reloadStyle(event.data);
    }
});
