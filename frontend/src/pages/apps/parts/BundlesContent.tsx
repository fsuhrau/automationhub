import React, {useState} from 'react';
import {Alert, Button, Card, CardContent, InputAdornment, LinearProgress, Stack, TextField, Typography} from '@mui/material';
import {CloudUploadRounded, SearchRounded} from '@mui/icons-material';
import {Loaded} from '../../../hooks/useLoader';
import {errorMessage} from '../../../lib/errorMessage';
import {IAppBinaryData} from '../../../types/app';
import {matches} from '../bundlesModel';
import {RowContext} from './BundleRow';
import {BundlesTable} from './BundlesTable';

interface Empty {
    accept: string,
    onChoose: () => void,
}

const NoBundles: React.FC<Empty> = ({accept, onChoose}) =>
    <Stack alignItems="center" spacing={1.5} sx={{py: 6, textAlign: 'center'}}>
        <CloudUploadRounded color="disabled" sx={{fontSize: 48}}/>
        <Typography variant="subtitle1">No bundles yet</Typography>
        <Typography variant="body2" color="text.secondary">
            Drop {accept} files here, or choose them. Test runs and live control install a bundle on the device.
        </Typography>
        <Button variant="outlined" onClick={onChoose}>Choose files</Button>
    </Stack>;

const countText = (shown: number, all: number): string => {
    if (shown !== all) return `${shown} of ${all}`;
    return `${all} ${all === 1 ? 'bundle' : 'bundles'}`;
};

interface SearchBarProps {
    search: string,
    shown: number,
    all: number,
    onSearch: (text: string) => void,
}

const SearchBar: React.FC<SearchBarProps> = ({search, shown, all, onSearch}) =>
    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1.5}}>
        <TextField size="small" placeholder="Find by version, tag or #id" value={search}
            onChange={e => onSearch(e.target.value)} sx={{width: {xs: '100%', sm: 300}}}
            slotProps={{
                       input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                       htmlInput: {'aria-label': 'Find a bundle'},
                   }}/>
        <Typography variant="body2" color="text.secondary" sx={{ml: 'auto'}}>{countText(shown, all)}</Typography>
    </Stack>;

interface BundleListProps {
    bundles: IAppBinaryData[],
    ctx: Omit<RowContext, 'onSearch'>,
}

const BundleList: React.FC<BundleListProps> = ({bundles, ctx}) => {
    const [search, setSearch] = useState('');
    const visible = bundles.filter(b => matches(b, search));
    return <>
        <SearchBar search={search} shown={visible.length} all={bundles.length} onSearch={setSearch}/>
        {visible.length === 0
            ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>No bundle matches.</Typography>
            : <BundlesTable bundles={visible} ctx={{...ctx, onSearch: setSearch}}/>}
    </>;
};

interface BundlesContentProps extends Empty {
    loader: Loaded<IAppBinaryData[]>,
    ctx: Omit<RowContext, 'onSearch' | 'newestId'>,
}

// BundlesContent is the card of the list: loading, failed, empty or the bundles
export const BundlesContent: React.FC<BundlesContentProps> = ({loader, accept, onChoose, ctx}) => {
    const {data, error, retry} = loader;
    return <Card variant="outlined">
        <CardContent>
            {error ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={retry}>Retry</Button>}>{errorMessage(error)}</Alert> : null}
            {!error && data === null && <LinearProgress/>}
            {!error && data?.length === 0 && <NoBundles accept={accept} onChoose={onChoose}/>}
            {!error && data && data.length > 0 && <BundleList bundles={data} ctx={{...ctx, newestId: data[0].id}}/>}
        </CardContent>
    </Card>;
};
