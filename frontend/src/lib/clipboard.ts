// copyText writes text to the clipboard. It never rejects: the result says
// whether it worked (the clipboard is missing outside secure contexts and
// the browser may deny the write).
export const copyText = async (text: string, clipboard: Pick<Clipboard, 'writeText'> | undefined = navigator.clipboard): Promise<boolean> => {
    if (!clipboard) return false;
    try {
        await clipboard.writeText(text);
        return true;
    } catch {
        return false;
    }
};
