var dropZones = document.getElementsByClassName('submit-form');
for (var i = 0; i < dropZones.length; i++) {
    var dropZone = dropZones[i];
    const hoverClassName = "hover";

    // Handle drag* events to handle style
    // Add the css you want when the class "hover" is present
    dropZone.addEventListener("dragenter", function (e) {
        e.preventDefault();
        dropZone.classList.add(hoverClassName);
    });

    dropZone.addEventListener("dragover", function (e) {
        e.preventDefault();
        dropZone.classList.add(hoverClassName);
    });

    dropZone.addEventListener("dragleave", function (e) {
        e.preventDefault();
        dropZone.classList.remove(hoverClassName);
    });

    
    dropZone.addEventListener("drop", function (e) {
        e.preventDefault();
        dropZone.classList.remove(hoverClassName);

        const files = Array.from(e.dataTransfer.files);
        uploadImage(e, files);        
    });

    dropZone.addEventListener("paste", async function (e) {
        if (!e.clipboardData.getData("text")) {
            e.preventDefault();        
            uploadImage(e, await getFilesAsync(e.clipboardData));        
        }
    });
}



async function getFilesAsync(dataTransfer) {
    const files = [];
    for (let i = 0; i < dataTransfer.items.length; i++) {
        const item = dataTransfer.items[i];
        if (item.kind === "file") {            
            const file = item.getAsFile();
            if (file) {
                files.push(file);
            }
        }
    }

    return files;
}

function picChange(e) {
    var fileInput = e.target.files;
    uploadImage(e, fileInput);
    return false;
}

function uploadImage(e, files) {
    if (files.length > 0) {
        const data = new FormData();
        for (const file of files) {
            data.append('file', file);
        }
        var imgelm = e.target.form.querySelector('img.uploaded-image');
        imgelm.src = '/static/img/loading.gif';
        imgelm.style.display = 'block';
        fetch('/upload', {
            method: 'POST',
            body: data
        }).then(function (response) {
            if (response.status == 200) {
                response.json().then(function(data) {
                   e.target.form.querySelector('input.image-url').value = data.url;
                   e.target.form.querySelector('img.uploaded-image').src = data.url;
                });                            
            } else {
                throw "None 200 response " + response.status;
            }
        })
        .catch(function (error) {
            imgelm.style.display = 'none';
            var snackbarContainer = document.querySelector('#error-snackbar');
            var data = {message: 'Error Uploading file'};
            snackbarContainer.MaterialSnackbar.showSnackbar(data);
        });
        return;
    }
}

// Returns a promise with all the files of the directory hierarchy
function readEntryContentAsync(entry) {
    return new Promise((resolve, reject) => {
        let reading = 0;
        const contents = [];

        readEntry(entry);

        function readEntry(entry) {
            if (isFile(entry)) {
                reading++;
                entry.file(file => {
                    reading--;
                    contents.push(file);

                    if (reading === 0) {
                        resolve(contents);
                    }
                });
            } else if (isDirectory(entry)) {
                readReaderContent(entry.createReader());
            }
        }

        function readReaderContent(reader) {
            reading++;

            reader.readEntries(function (entries) {
                reading--;
                for (const entry of entries) {
                    readEntry(entry);
                }

                if (reading === 0) {
                    resolve(contents);
                }
            });
        }
    });
}

function isDirectory(entry) {
    if (entry) {
        return entry.isDirectory;
    }
    return false;
}

function isFile(entry) {
    if (entry) {
        return entry.isFile;
    }
    return false;
}

