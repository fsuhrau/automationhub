const hex = (b: number): string => b.toString(16).padStart(2, '0');

// avc1.PPCCLL from the first SPS NAL unit of an Annex B buffer.
export const h264CodecString = (data: Uint8Array): string | null => {
    for (let i = 0; i + 3 < data.length; i++) {
        if (data[i] !== 0 || data[i + 1] !== 0 || data[i + 2] !== 1) continue;
        const nal = i + 3;
        if ((data[nal] & 0x1f) === 7 && nal + 3 < data.length) {
            return `avc1.${hex(data[nal + 1])}${hex(data[nal + 2])}${hex(data[nal + 3])}`;
        }
        i = nal - 1;
    }
    return null;
};
