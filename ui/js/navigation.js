document.addEventListener("click", async (e) => {
    const link = e.target.closest("[data-transition]");
    if (!link) return;

    e.preventDefault();

    const url = link.href;

    const app = document.querySelector("#app");

    app.style.transition = "opacity 200ms ease";
    app.style.opacity = "0";

    await new Promise(r => setTimeout(r, 200));

    const res = await fetch(url);
    const html = await res.text();

    const doc = new DOMParser().parseFromString(html, "text/html");

    const newContent = doc.querySelector("#app");

    if (!newContent) {
        console.error("No #app found in response");
        return;
    }

    app.innerHTML = newContent.innerHTML;

    window.scrollTo(0, 0);

    app.style.opacity = "1";
});