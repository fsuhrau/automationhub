import React from 'react';
import {DatePicker, LocalizationProvider} from '@mui/x-date-pickers';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import {Dayjs} from 'dayjs';

export interface ExpiryPickerProps {
    value: Dayjs | null,
    onChange: (value: Dayjs | null) => void,
    // the first date that can be picked; today when not given
    minDate?: Dayjs,
}

// ExpiryPicker picks when a token stops working; empty means never
const ExpiryPicker: React.FC<ExpiryPickerProps> = ({value, onChange, minDate}) =>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
        <DatePicker label="Expires (optional)" value={value} onChange={onChange} disablePast={!minDate} minDate={minDate}
            slotProps={{textField: {size: 'small', helperText: 'Empty: until revoked'}, field: {clearable: true}}}/>
    </LocalizationProvider>;

export default ExpiryPicker;
