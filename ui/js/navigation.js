  
document.addEventListener("click", async (e) => {
    
    console.log("nav start" );
    
    const link = e.target.closest("[data-transition]");
    if (!link) return;

    console.log("nav click", link.href);

    e.preventDefault();

    const url = link.href;

    const app = document.querySelector("#app");

    app.style.transition = "opacity 200ms ease";
    app.style.opacity = "0";

    console.log("before delay" );
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

    console.log("before init" );
    initPage();

    window.scrollTo(0, 0);

    app.style.opacity = "1";
});