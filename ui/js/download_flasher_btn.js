function initDownloadButtons() {

    document.querySelectorAll(".download-wrapper").forEach(wrapper => {

        const items = wrapper.querySelectorAll(".download-item");

        items.forEach((item, index) => {
            item.style.setProperty("--delay", index);
        });


        const downloadBtn = wrapper.querySelector(".btn-primary");
        const downloadPanel = wrapper.querySelector(".download-panel");


        if (!downloadBtn || !downloadPanel) {
            return;
        }


        downloadBtn.onclick = () => {
            downloadPanel.classList.toggle("open");
        };

    });

}
 