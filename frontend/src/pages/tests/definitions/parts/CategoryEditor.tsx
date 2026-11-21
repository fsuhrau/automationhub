import React, {useState} from 'react';
import {Box, Button, Chip, Stack, TextField} from '@mui/material';
import {mergeCategories} from '../../testEditModel';

interface CategoryEditorProps {
    categories: string[],
    onChange: (categories: string[]) => void,
    error?: string,
}

// CategoryEditor names the Unity categories to run; Enter adds, commas
// separate several
export const CategoryEditor: React.FC<CategoryEditorProps> = ({categories, onChange, error}) => {
    const [text, setText] = useState('');
    const add = (): void => {
        onChange(mergeCategories(categories, text));
        setText('');
    };
    return <Box sx={{mt: 1.5}}>
        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} sx={{mb: 1}}>
            {categories.map(c => <Chip key={c} size="small" label={c} onDelete={() => onChange(categories.filter(x => x !== c))}/>)}
        </Stack>
        <Stack direction="row" spacing={1}>
            <TextField size="small" label="Category" value={text} onChange={e => setText(e.target.value)}
                error={!!error} helperText={error ?? 'Enter adds it; commas separate several.'}
                onKeyDown={e => {
                           if (e.key === 'Enter') {
                               e.preventDefault();
                               add();
                           }
                       }}/>
            <Box><Button variant="outlined" onClick={add} disabled={!text.trim()}>Add</Button></Box>
        </Stack>
    </Box>;
};
