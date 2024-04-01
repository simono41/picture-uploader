document.addEventListener('DOMContentLoaded', (event) => {
    document.getElementById('copyButton').addEventListener('click', kopiereURL);
});

function kopiereURL() {
    var copyText = document.getElementById("imageURL");
    copyText.select();
    copyText.setSelectionRange(0, 99999); // Für mobile Geräte

    navigator.clipboard.writeText(copyText.value).then(function() {
        console.log('Kopieren in die Zwischenablage erfolgreich.');
    }, function(err) {
        console.error('Fehler beim Kopieren in die Zwischenablage: ', err);
    });
}
