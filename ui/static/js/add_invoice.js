document.addEventListener("DOMContentLoaded", () => {
        const dateInput = document.getElementById("dateInput");
        if (dateInput) {
            const d = new Date();
            d.setMonth(d.getMonth() - 1);
            dateInput.value = d.toISOString().split('T')[0];
        }
    });
