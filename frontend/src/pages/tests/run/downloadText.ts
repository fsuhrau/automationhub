// downloadText saves text as a file through the browser
export const downloadText = (fileName: string, text: string): void => {
    const url = URL.createObjectURL(new Blob([text], {type: 'text/plain'}));
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
};
