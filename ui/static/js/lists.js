document.addEventListener("DOMContentLoaded", () => {
    
    const rows = document.querySelectorAll("tr.clickable-row");
    if (rows.length > 0) {
        rows.forEach(row => {
            row.addEventListener("click", () => {
                const selection = window.getSelection();
                if (selection.toString().length === 0) {
                    window.location.href = row.dataset.href;
                }
            });
        });
    }

});
