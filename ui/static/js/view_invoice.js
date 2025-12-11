document.addEventListener("DOMContentLoaded", function() {
    const nettoEl = document.querySelector('.val-netto');
    const podatekEl = document.querySelector('.val-podatek');
    const bruttoEl = document.querySelector('.val-brutto');

    function formatMoney(value) {
        return value.toFixed(2) + ' PLN';
    }

    if (nettoEl && podatekEl && bruttoEl) {
        const netto = parseFloat(nettoEl.innerText) || 0;
        const podatek = parseFloat(podatekEl.innerText) || 0;
        const sum = netto + podatek;
        nettoEl.innerText = formatMoney(netto);
        podatekEl.innerText = formatMoney(podatek);
        bruttoEl.innerText = formatMoney(sum);
    }
});
