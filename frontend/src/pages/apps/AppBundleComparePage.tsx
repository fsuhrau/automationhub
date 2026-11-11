import React, {useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Avatar,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    FormControl,
    FormControlLabel,
    IconButton,
    InputLabel,
    LinearProgress,
    Link,
    MenuItem,
    Select,
    Stack,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Tooltip,
    Typography,
} from '@mui/material';
import {ArchiveOutlined, ArrowBackRounded, SwapHorizRounded} from '@mui/icons-material';
import {Link as RouterLink, useSearchParams} from 'react-router-dom';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {IAppBinaryData, IAppBundleDetail} from '../../types/app';
import {bundleIconUrl, getAppBundle, getAppBundles} from '../../services/app.service';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {formatDateTime} from '../../components/DateTime';
import {byteFormat} from '../tests/value_formatter';
import {ago} from './bundlesModel';
import {
    Change,
    compareBundles,
    hideEmpty,
    ListRow,
    onlyChanges,
    percent,
    previousOf,
    Section,
    signedBytes,
    signedCount,
    SizeRow,
    ValueRow,
} from './bundleCompareModel';

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const mono = {fontFamily: 'monospace', fontSize: '0.8125rem', wordBreak: 'break-all'} as const;

const CHANGE: Record<Change, { label: string, color: 'default' | 'success' | 'error' | 'warning' }> = {
    same: {label: 'same', color: 'default'},
    changed: {label: 'changed', color: 'warning'},
    added: {label: 'added', color: 'success'},
    removed: {label: 'removed', color: 'error'},
};

// edgeColor marks a row by how it changed
const edgeColor = (change: Change) => change === 'same' ? 'transparent'
    : change === 'added' ? 'success.main' : change === 'removed' ? 'error.main' : 'warning.main';

const ChangeMark: React.FC<{ change: Change }> = ({change}) =>
    change === 'same' ? <Typography variant="caption" color="text.disabled">same</Typography>
        : <Chip size="small" color={CHANGE[change].color} variant="outlined" label={CHANGE[change].label}/>;

const bundleLabel = (b: IAppBinaryData) => {
    const version = b.info?.version || b.version || b.name;
    const build = b.info?.build && b.info.build !== version ? ` (${b.info.build})` : '';
    return `${version || '–'}${build} · #${b.id}`;
};

const Delta: React.FC<{ row: ValueRow }> = ({row}) => {
    if (row.delta === undefined || row.delta === 0) return null;
    const text = row.unit === 'bytes' ? signedBytes(row.delta) : signedCount(row.delta);
    const share = percent(row.from, row.delta);
    return <Typography component="span" variant="caption" sx={{ml: 1, fontWeight: 600, color: row.delta > 0 ? 'warning.main' : 'success.main'}}>
        {text}{share && ` (${share})`}
    </Typography>;
};

const ValueCells: React.FC<{ row: ValueRow }> = ({row}) => <>
    <TableCell sx={{verticalAlign: 'top', color: row.a === undefined ? 'text.disabled' : undefined, wordBreak: 'break-word'}}>
        {row.a ?? '–'}
    </TableCell>
    <TableCell sx={{verticalAlign: 'top', color: row.b === undefined ? 'text.disabled' : undefined, wordBreak: 'break-word',
        fontWeight: row.change === 'same' ? undefined : 600}}>
        {row.b ?? '–'}<Delta row={row}/>
    </TableCell>
</>;

const ListCells: React.FC<{ row: ListRow }> = ({row}) => {
    const [all, setAll] = useState(false);
    const max = 30;
    const chips = (items: string[], color: 'success' | 'error', sign: string) => {
        const shown = all ? items : items.slice(0, max);
        return <>
            {shown.map(i => <Chip key={sign + i} size="small" color={color} variant="outlined" label={`${sign} ${i}`}
                                  sx={{maxWidth: '100%', '& .MuiChip-label': {overflow: 'hidden', textOverflow: 'ellipsis'}}}/>)}
            {items.length > shown.length && <Chip size="small" label={`+${items.length - shown.length} more`} onClick={() => setAll(true)}/>}
        </>;
    };
    return <TableCell colSpan={2} sx={{verticalAlign: 'top'}}>
        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap alignItems="center">
            {chips(row.added, 'success', '+')}
            {chips(row.removed, 'error', '−')}
            <Typography variant="caption" color="text.secondary" sx={{ml: 0.5}}>
                {row.change === 'same' ? `${row.same} the same in both` : row.same > 0 ? `${row.same} unchanged` : ''}
            </Typography>
        </Stack>
    </TableCell>;
};

const SectionCard: React.FC<{ section: Section }> = ({section}) =>
    <Card variant="outlined">
        <CardContent>
            <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>{section.title}</Typography>
            <TableContainer sx={{mx: -2, width: 'auto'}}>
                <Table size="small" aria-label={section.title} sx={{tableLayout: 'fixed'}}>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{pl: 2, width: {xs: '28%', md: '20%'}}}>Attribute</TableCell>
                            <TableCell>From</TableCell>
                            <TableCell>To</TableCell>
                            <TableCell sx={{width: 96, display: {xs: 'none', sm: 'table-cell'}}}/>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {section.rows.map(r => <TableRow key={r.label} sx={{
                            '& > td:first-of-type': {
                                borderLeft: 3, borderLeftColor: edgeColor(r.change),
                            },
                        }}>
                            <TableCell sx={{pl: 2, verticalAlign: 'top', color: 'text.secondary'}}>{r.label}</TableCell>
                            {r.kind === 'value' ? <ValueCells row={r}/> : <ListCells row={r}/>}
                            <TableCell sx={{verticalAlign: 'top', display: {xs: 'none', sm: 'table-cell'}}}><ChangeMark change={r.change}/></TableCell>
                        </TableRow>)}
                    </TableBody>
                </Table>
            </TableContainer>
        </CardContent>
    </Card>;

// SizesCard shows how the size of each part or file changed, biggest change first
const SizesCard: React.FC<{ title: string, subtitle: string, rows: SizeRow[], onlyChanged: boolean, missing: string }> =
    ({title, subtitle, rows, onlyChanged, missing}) => {
        const [all, setAll] = useState(false);
        const filtered = onlyChanged ? rows.filter(r => r.change !== 'same') : rows;
        const shown = all ? filtered : filtered.slice(0, 15);
        const max = Math.max(1, ...filtered.map(r => Math.abs(r.delta ?? 0)));
        if (rows.length === 0) return null;
        return <Card variant="outlined">
            <CardContent>
                <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                <Typography variant="body2" color="text.secondary" sx={{mb: 1}}>{subtitle}</Typography>
                {filtered.length === 0
                    ? <Typography variant="body2" color="text.secondary">No size changed.</Typography>
                    : <TableContainer sx={{mx: -2, width: 'auto'}}>
                        <Table size="small" aria-label={title}>
                            <TableHead>
                                <TableRow>
                                    <TableCell sx={{pl: 2}}>Path</TableCell>
                                    <TableCell align="right" sx={{display: {xs: 'none', sm: 'table-cell'}}}>From</TableCell>
                                    <TableCell align="right" sx={{display: {xs: 'none', sm: 'table-cell'}}}>To</TableCell>
                                    <TableCell align="right">Change</TableCell>
                                    <TableCell sx={{width: '22%', display: {xs: 'none', md: 'table-cell'}}}/>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {shown.map(r => <TableRow key={r.path}>
                                    <TableCell sx={{pl: 2, ...mono}}>{r.path}</TableCell>
                                    <TableCell align="right" sx={{whiteSpace: 'nowrap', display: {xs: 'none', sm: 'table-cell'}}}>
                                        {r.a === undefined ? <Typography component="span" variant="caption" color="text.disabled">{missing}</Typography> : byteFormat(r.a)}
                                    </TableCell>
                                    <TableCell align="right" sx={{whiteSpace: 'nowrap', display: {xs: 'none', sm: 'table-cell'}}}>
                                        {r.b === undefined ? <Typography component="span" variant="caption" color="text.disabled">{missing}</Typography> : byteFormat(r.b)}
                                    </TableCell>
                                    <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                                        {r.delta === undefined
                                            ? <Typography component="span" variant="caption" color="text.disabled">unknown</Typography>
                                            : <>
                                                <Typography component="span" variant="body2" sx={{fontWeight: 600,
                                                    color: r.delta > 0 ? 'warning.main' : r.delta < 0 ? 'success.main' : 'text.secondary'}}>
                                                    {signedBytes(r.delta)}
                                                </Typography>
                                                {r.a !== undefined && r.b !== undefined && percent(r.a, r.delta) &&
                                                    <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 0.75}}>{percent(r.a, r.delta)}</Typography>}
                                            </>}
                                    </TableCell>
                                    <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>
                                        {/* growth to the right, shrinking to the left of the middle */}
                                        <Box sx={{display: 'flex', height: 8}} aria-hidden>
                                            <Box sx={{flex: 1, display: 'flex', justifyContent: 'flex-end'}}>
                                                {r.delta !== undefined && r.delta < 0 && <Box sx={{width: `${Math.abs(r.delta) / max * 100}%`, bgcolor: 'success.main', borderRadius: '4px 0 0 4px'}}/>}
                                            </Box>
                                            <Box sx={{width: '1px', bgcolor: 'divider'}}/>
                                            <Box sx={{flex: 1}}>
                                                {r.delta !== undefined && r.delta > 0 && <Box sx={{width: `${r.delta / max * 100}%`, height: '100%', bgcolor: 'warning.main', borderRadius: '0 4px 4px 0'}}/>}
                                            </Box>
                                        </Box>
                                    </TableCell>
                                </TableRow>)}
                            </TableBody>
                        </Table>
                    </TableContainer>}
                {filtered.length > shown.length && <Button size="small" sx={{mt: 1}} onClick={() => setAll(true)}>Show all {filtered.length}</Button>}
            </CardContent>
        </Card>;
    };

const BundleHeader: React.FC<{ side: string, bundle: IAppBundleDetail, iconUrl?: string, appBase: string, platform?: number, now: number }> =
    ({side, bundle, iconUrl, appBase, platform, now}) => {
        const [failed, setFailed] = useState(false);
        return <Card variant="outlined" sx={{flex: 1, minWidth: 0}}>
            <CardContent>
                <Typography variant="overline" color="text.secondary">{side}</Typography>
                <Stack direction="row" spacing={1.5} alignItems="center">
                    {iconUrl && !failed
                        ? <Avatar variant="rounded" src={iconUrl} alt="" imgProps={{onError: () => setFailed(true)}}
                                  sx={{width: 40, height: 40, border: 1, borderColor: 'divider'}}/>
                        : <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', width: 40, height: 40, borderRadius: 1, border: 1, borderColor: 'divider'}}>
                            {platform !== undefined && <PlatformTypeIcon platformType={platform}/>}
                        </Box>}
                    <Box sx={{minWidth: 0}}>
                        <Link component={RouterLink} to={`${appBase}/bundle/${bundle.id}`} variant="subtitle1" sx={{fontWeight: 600, wordBreak: 'break-word'}}>
                            {bundleLabel(bundle)}
                        </Link>
                        <Typography variant="body2" color="text.secondary">
                            <Tooltip title={formatDateTime(bundle.createdAt)}><span>uploaded {ago(bundle.createdAt, now)}</span></Tooltip>
                            {' · '}{bundle.runCount} {bundle.runCount === 1 ? 'run' : 'runs'}
                        </Typography>
                        {bundle.archivedAt && <Chip size="small" variant="outlined" icon={<ArchiveOutlined/>} label="archived" sx={{mt: 0.5}}/>}
                    </Box>
                </Stack>
                {!bundle.info && <Typography variant="caption" color="warning.main" sx={{display: 'block', mt: 1}}>
                    Uploaded before the master read the details of a file; only the basics can be compared.
                </Typography>}
            </CardContent>
        </Card>;
    };

// AppBundleComparePage puts two bundles of the app side by side and shows
// what changed from the first to the second: attributes, lists like the
// permissions, and where the size went. Archived bundles compare too.
const AppBundleComparePage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const app = state.apps?.find(a => a.id === appId);
    const appBase = `/project/${projectIdentifier}/app:${appId}`;
    const [params, setParams] = useSearchParams();
    const now = Date.now();

    const [bundles, setBundles] = useState<IAppBinaryData[] | null>(null);
    const [details, setDetails] = useState<Record<number, IAppBundleDetail>>({});
    const [error, setError] = useState<string | null>(null);
    const [changesOnly, setChangesOnly] = useState(true);

    useEffect(() => {
        if (appId === null) return;
        getAppBundles(projectIdentifier, appId).then(list => setBundles([...list].sort((x, y) => y.id - x.id)))
            .catch(e => setError(messageOf(e)));
    }, [projectIdentifier, appId]);

    // without ids, compare the newest bundle to the one before it; with only
    // the second, compare it to the one before it (the oldest to the next)
    const bId = Number(params.get('b')) || bundles?.[0]?.id;
    const aId = Number(params.get('a')) || (bundles && bId ? (previousOf(bundles, bId) ?? bundles.find(x => x.id !== bId))?.id : undefined);

    useEffect(() => {
        if (appId === null) return;
        for (const id of [aId, bId]) {
            if (!id || details[id]) continue;
            getAppBundle(projectIdentifier, appId, id).then(d => setDetails(m => ({...m, [id]: d})))
                .catch(e => setError(messageOf(e)));
        }
    }, [projectIdentifier, appId, aId, bId, details]);

    const a = aId ? details[aId] : undefined;
    const b = bId ? details[bId] : undefined;
    const comparison = useMemo(() => a && b ? compareBundles(a, b) : undefined, [a, b]);
    const sections = useMemo(() => comparison ? (changesOnly ? onlyChanges(comparison.sections) : hideEmpty(comparison.sections)) : [],
        [comparison, changesOnly]);

    const pick = (which: 'a' | 'b', id: number) => {
        const next = {a: String(aId ?? ''), b: String(bId ?? ''), [which]: String(id)};
        setParams(next, {replace: true});
    };
    const swap = () => setParams({a: String(bId ?? ''), b: String(aId ?? '')}, {replace: true});
    const icon = (d?: IAppBundleDetail) => d?.info?.icon && appId !== null ? bundleIconUrl(projectIdentifier, appId, d.id) : undefined;

    const picker = (which: 'a' | 'b', label: string, value?: number) =>
        <FormControl size="small" sx={{flex: 1, minWidth: 0}}>
            <InputLabel id={`compare-${which}`}>{label}</InputLabel>
            <Select labelId={`compare-${which}`} label={label} value={value && bundles?.some(x => x.id === value) ? value : ''}
                    onChange={e => pick(which, Number(e.target.value))}>
                {(bundles ?? []).map(x => <MenuItem key={x.id} value={x.id}>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{minWidth: 0}}>
                        <span>{bundleLabel(x)}</span>
                        <Typography component="span" variant="caption" color="text.secondary">{ago(x.createdAt, now)}</Typography>
                        {x.tags && <Typography component="span" variant="caption" color="text.secondary">{x.tags}</Typography>}
                        {x.archivedAt && <ArchiveOutlined fontSize="inherit" color="disabled" aria-label="archived"/>}
                    </Stack>
                </MenuItem>)}
            </Select>
        </FormControl>;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5">Compare bundles</Typography>
                    <Typography variant="body2" color="text.secondary">
                        What changed from one build of {app?.name ?? 'the app'} to another. Archived bundles keep their details and compare too.
                    </Typography>
                </Box>
                <Button sx={{ml: {md: 'auto'}}} startIcon={<ArrowBackRounded/>} component={RouterLink} to={`${appBase}/bundles`}>All bundles</Button>
            </Stack>

            {error && <Alert severity="error" sx={{mb: 2}} onClose={() => setError(null)}>{error}</Alert>}

            <Card variant="outlined" sx={{mb: 2}}>
                <CardContent>
                    {bundles === null ? <LinearProgress/>
                        : bundles.length < 2 ? <Typography variant="body2" color="text.secondary">The app needs two bundles to compare.</Typography>
                            : <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'center'}}>
                                {picker('a', 'From', aId)}
                                <Tooltip title="Swap">
                                    <IconButton aria-label="Swap the bundles" onClick={swap} sx={{alignSelf: 'center'}}><SwapHorizRounded/></IconButton>
                                </Tooltip>
                                {picker('b', 'To', bId)}
                            </Stack>}
                </CardContent>
            </Card>

            {bundles && bundles.length >= 2 && (!a || !b) && !error && <LinearProgress/>}

            {a && b && comparison && <>
                <Stack direction={{xs: 'column', md: 'row'}} spacing={2} sx={{mb: 2}}>
                    <BundleHeader side="From" bundle={a} iconUrl={icon(a)} appBase={appBase} platform={app?.platform} now={now}/>
                    <BundleHeader side="To" bundle={b} iconUrl={icon(b)} appBase={appBase} platform={app?.platform} now={now}/>
                </Stack>

                <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 2}}>
                    <Typography variant="body1" sx={{fontWeight: 600}} aria-live="polite">
                        {a.id === b.id ? 'The same bundle on both sides.'
                            : comparison.changes === 0 ? 'No attribute changed.'
                                : `${comparison.changes} ${comparison.changes === 1 ? 'attribute changed' : 'attributes changed'}`}
                    </Typography>
                    {a.id > b.id && <Chip size="small" color="info" variant="outlined" label="From is the newer bundle"/>}
                    <FormControlLabel sx={{ml: 'auto'}} control={<Switch checked={changesOnly} onChange={e => setChangesOnly(e.target.checked)}/>}
                                      label="Only changes"/>
                </Stack>

                <Stack spacing={2}>
                    {sections.map(s => <SectionCard key={s.title} section={s}/>)}
                    <SizesCard title="Size by part" subtitle="Installed size of each part of the bundle"
                               rows={comparison.parts} onlyChanged={changesOnly} missing="not there"/>
                    <SizesCard title="Largest files" subtitle="The largest files of either bundle; a file missing on one side may just be smaller there"
                               rows={comparison.files} onlyChanged={changesOnly} missing="not among largest"/>
                </Stack>
            </>}
        </Box>
    );
};

export default AppBundleComparePage;
