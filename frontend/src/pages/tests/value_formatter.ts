export const kFormat = (value: number): string => {
    if (value < 1000) {
        return value.toFixed(2)
    }
    return (value / 1000).toFixed(2) + "k"
}

export const fixedTwoFormat = (value: number): string => {
    return value.toFixed(2)
}

export const byteFormat = (bytes: number): string => {
    const KB = 1024;
    const MB = KB * 1024;
    const GB = MB * 1024;

    if (bytes >= GB) {
        return `${ (bytes / GB).toFixed(2) }GB`;
    }

    if (bytes >= MB) {
        return `${ (bytes / MB).toFixed(2) }MB`;
    }

    if (bytes >= KB) {
        return `${ (bytes / KB).toFixed(2) }KB`;
    }

    return `${ bytes }B`;
};
