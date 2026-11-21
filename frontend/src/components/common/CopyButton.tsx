import React, {useState} from 'react';
import {IconButton, Tooltip} from '@mui/material';
import {CheckRounded, ContentCopyRounded} from '@mui/icons-material';
import {copyText} from '../../lib/clipboard';

const COPIED_MS = 1500;

export interface CopyButtonProps {
    text: string,
    label: string,
}

// CopyButton copies text and shows a check for a moment
export const CopyButton: React.FC<CopyButtonProps> = ({text, label}) => {
    const [copied, setCopied] = useState(false);
    const copy = async (): Promise<void> => {
        if (!await copyText(text)) return;
        setCopied(true);
        setTimeout(() => setCopied(false), COPIED_MS);
    };
    return <Tooltip title={copied ? 'Copied' : label}>
        <IconButton size="small" aria-label={label} onClick={copy}>
            {copied ? <CheckRounded fontSize="small" color="success"/> : <ContentCopyRounded fontSize="small"/>}
        </IconButton>
    </Tooltip>;
};
