import React from 'react';
import {Box, Chip, IconButton, Link, Stack, TableCell, TableRow, Tooltip, Typography, Button} from '@mui/material';
import {ArchiveOutlined, DownloadRounded, EditRounded, MoreVertRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {formatDateTime} from '../../../components/DateTime';
import {IAppBinaryData} from '../../../types/app';
import {byteFormat} from '../../tests/value_formatter';
import {ago, downloadUrl, isInstallable, parseTags} from '../bundlesModel';
import {BundleIcon} from './BundleIcon';
import {hideBelow} from './tableStyle';

// what a row of the table needs to know about the table around it
export interface RowContext {
    // the bundle uploaded last
    newestId: number,
    // the bundle that was just uploaded
    fresh: number | null,
    now: number,
    detailUrl: (id: number) => string,
    iconUrl: (b: IAppBinaryData) => string | undefined,
    onSearch: (text: string) => void,
    onTags: (b: IAppBinaryData) => void,
    onMenu: (anchor: HTMLElement, b: IAppBinaryData) => void,
}

const engineText = (b: IAppBinaryData): string => {
    const engine = b.info?.engine;
    return engine ? ` · ${[engine.name, engine.version].filter(Boolean).join(' ')}` : '';
};

const VersionCell: React.FC<{ bundle: IAppBinaryData, ctx: RowContext }> = ({bundle: b, ctx}) =>
    <TableCell sx={{pl: 2, maxWidth: 420}}>
        <Stack direction="row" spacing={1.5} alignItems="center">
            <BundleIcon src={ctx.iconUrl(b)} size={32} radius={1.5}/>
            <Box sx={{minWidth: 0}}>
                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
                    <Link component={RouterLink} to={ctx.detailUrl(b.id)} underline="hover" variant="body2"
                        sx={{fontWeight: 600, wordBreak: 'break-all'}}>{b.version || '–'}</Link>
                    {b.info?.build && b.info.build !== b.version && <Typography variant="caption" color="text.secondary">build {b.info.build}</Typography>}
                    {b.id === ctx.newestId && <Chip size="small" variant="outlined" label="newest"/>}
                    {b.id === ctx.fresh && <Chip size="small" color="success" label="just uploaded"/>}
                    {b.archivedAt && <Tooltip title={`Archived ${formatDateTime(b.archivedAt)}: the file is gone, the details stay for comparing runs`}>
                        <Chip size="small" variant="outlined" icon={<ArchiveOutlined/>} label="archived"/>
                    </Tooltip>}
                </Stack>
                <Typography variant="caption" color="text.secondary">{b.info?.displayName || b.name} · #{b.id}{engineText(b)}</Typography>
            </Box>
        </Stack>
    </TableCell>;

const TagsCell: React.FC<{ bundle: IAppBinaryData, ctx: RowContext }> = ({bundle: b, ctx}) => {
    const tags = parseTags(b.tags);
    const edit = (): void => ctx.onTags(b);
    return <TableCell>
        <Stack direction="row" spacing={0.5} alignItems="center" flexWrap="wrap" useFlexGap={true}>
            {tags.map(t => <Chip key={t} size="small" label={t} onClick={() => ctx.onSearch(t)}/>)}
            {tags.length > 0
                ? <Tooltip title="Edit tags">
                    <IconButton size="small" aria-label={`Edit tags of #${b.id}`} onClick={edit}><EditRounded fontSize="inherit"/></IconButton>
                </Tooltip>
                : <Button size="small" variant="text" aria-label={`Add tags to #${b.id}`} onClick={edit}>Add tags</Button>}
        </Stack>
    </TableCell>;
};

const ActionsCell: React.FC<{ bundle: IAppBinaryData, ctx: RowContext }> = ({bundle: b, ctx}) =>
    <TableCell align="right" sx={{whiteSpace: 'nowrap', pr: 1}}>
        <Tooltip title={b.archivedAt ? 'Archived: the file is gone' : 'Download'}>
            <span><IconButton size="small" aria-label={`Download #${b.id}`} disabled={!isInstallable(b)} href={downloadUrl(b.appPath)} download={true}>
                <DownloadRounded fontSize="small"/>
            </IconButton></span>
        </Tooltip>
        <IconButton size="small" aria-label={`More for #${b.id}`} onClick={e => ctx.onMenu(e.currentTarget, b)}>
            <MoreVertRounded fontSize="small"/>
        </IconButton>
    </TableCell>;

// BundleRow is one build in the list
export const BundleRow: React.FC<{ bundle: IAppBinaryData, ctx: RowContext }> = ({bundle: b, ctx}) =>
    <TableRow selected={b.id === ctx.fresh}>
        <VersionCell bundle={b} ctx={ctx}/>
        <TagsCell bundle={b} ctx={ctx}/>
        <TableCell sx={{...hideBelow('md'), whiteSpace: 'nowrap'}}>
            <Tooltip title={formatDateTime(b.createdAt)}><span>{ago(b.createdAt, ctx.now)}</span></Tooltip>
        </TableCell>
        <TableCell align="right" sx={{...hideBelow('sm'), whiteSpace: 'nowrap'}}>{b.size ? byteFormat(b.size) : '–'}</TableCell>
        <TableCell align="right" sx={hideBelow('md')}>{b.runCount === undefined ? '–' : b.runCount}</TableCell>
        <ActionsCell bundle={b} ctx={ctx}/>
    </TableRow>;
