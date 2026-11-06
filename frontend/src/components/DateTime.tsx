import dayjs from 'dayjs';

export const DEFAULT_DATE_TIME_FORMAT = 'YYYY/MM/DD HH:mm:ss';

export const formatDateTime = (value: Date | string | number | null | undefined, format: string = DEFAULT_DATE_TIME_FORMAT): string => {
    if (value === null || value === undefined || value === '') {
        return '';
    }
    const date = dayjs(value);
    return date.isValid() ? date.format(format) : '';
};

const DateTime = ({value, format}: { value: Date | string | number | null | undefined, format?: string }) => (
    <>{formatDateTime(value, format)}</>
);

export default DateTime;
