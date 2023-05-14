function reloadStyle(filename) {
    let link = document.querySelector(`link[href^="/style/${filename}"]`);
    link.href = `/style/${filename}?reload=${new Date().getTime()}`;
}

window.addEventListener("load", () => {
    const eventSource = new EventSource("/event/style");

    eventSource.onmessage = event => {
        reloadStyle(event.data);
    }
});
