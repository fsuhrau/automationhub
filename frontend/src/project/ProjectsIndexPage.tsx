import React, {useEffect, useMemo, useState} from 'react';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    ButtonBase,
    Card,
    Chip,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    IconButton,
    InputAdornment,
    LinearProgress,
    Stack,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import CssBaseline from '@mui/material/CssBaseline';
import {AddRounded, SearchRounded, SettingsRounded} from '@mui/icons-material';
import IProject from './project';
import {createProject, getProjects} from './project.service';
import {HubStateActions} from '../application/HubState';
import PlatformTypeIcon from '../components/PlatformTypeIcon';
import TopBar from '../components/TopBar';
import {useHubState} from '../hooks/HubStateProvider';
import {lastProject, matchesProject, projectSlug, sortProjects} from './projectsModel';

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const NewProjectDialog: React.FC<{ open: boolean, onClose: () => void, onCreated: (p: IProject) => void }> = ({open, onClose, onCreated}) => {
    const [name, setName] = useState('');
    const [tried, setTried] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);
    useEffect(() => {
        if (!open) return;
        setName('');
        setTried(false);
        setError(null);
    }, [open]);

    const submit = () => {
        setTried(true);
        if (!name.trim()) return;
        setBusy(true);
        createProject({name: name.trim()} as IProject).then(onCreated).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    return <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs" aria-labelledby="new-project-title">
        <Box component="form" noValidate onSubmit={e => {
            e.preventDefault();
            submit();
        }}>
            <DialogTitle id="new-project-title">New project</DialogTitle>
            <DialogContent>
                <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                    A project holds a game's apps, tests and results. Devices and nodes are shared by all projects.
                </Typography>
                <TextField autoFocus fullWidth size="small" label="Name" placeholder="e.g. Heroes of History" value={name}
                           onChange={e => setName(e.target.value)} error={tried && !name.trim()}
                           helperText={tried && !name.trim()
                               ? 'A name is needed.'
                               : name.trim() ? <>Identifier: <b>{projectSlug(name)}</b> (in URLs and the API; it can't change)</> : ' '}/>
                {error && <Alert severity="error" sx={{mt: 2}}>{error}</Alert>}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button type="submit" variant="contained" disabled={busy}>Create project</Button>
            </DialogActions>
        </Box>
    </Dialog>;
};

// ProjectsIndexPage is the start: the projects with their apps, the last
// opened one first, and creating a new one
const ProjectsIndexPage: React.FC = () => {
    const {state, dispatch} = useHubState();
    const navigate = useNavigate();
    const [error, setError] = useState<string | null>(null);
    const [creating, setCreating] = useState(false);
    const [search, setSearch] = useState('');
    const last = useMemo(lastProject, []);

    const load = () => {
        setError(null);
        getProjects().then(projects => dispatch({type: HubStateActions.ProjectsUpdate, payload: projects}))
            .catch(e => setError(messageOf(e)));
    };
    useEffect(() => {
        if (state.projects === null) load();
    }, [state.projects]);

    const projects = useMemo(() => sortProjects(state.projects ?? [], last).filter(p => matchesProject(p, search)), [state.projects, last, search]);
    const total = state.projects?.length ?? 0;

    return (
        <>
            <CssBaseline enableColorScheme/>
            <TopBar/>
            <Box sx={{maxWidth: 1200, mx: 'auto', px: {xs: 2, sm: 3}, py: 4}}>
                <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap alignItems={{sm: 'center'}} sx={{mb: 3}}>
                    <Box>
                        <Typography variant="h5" component="h1">Projects</Typography>
                        <Typography variant="body2" color="text.secondary">Pick a project to see its apps, tests and devices.</Typography>
                    </Box>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {sm: 'auto'}}}>
                        {total > 6 && <TextField size="small" placeholder="Find a project or app" value={search} onChange={e => setSearch(e.target.value)}
                                                 slotProps={{
                                                     input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                                     htmlInput: {'aria-label': 'Find a project'},
                                                 }}/>}
                        <Button variant="contained" startIcon={<AddRounded/>} onClick={() => setCreating(true)}>New project</Button>
                    </Stack>
                </Stack>

                {error
                    ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{error}</Alert>
                    : state.projects === null
                        ? <LinearProgress/>
                        : total === 0
                            ? <Card variant="outlined" sx={{p: 4, textAlign: 'center'}}>
                                <Typography variant="subtitle1" gutterBottom>No projects yet</Typography>
                                <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                                    Create one for your game; then add its apps and tests.
                                </Typography>
                                <Button variant="contained" startIcon={<AddRounded/>} onClick={() => setCreating(true)}>New project</Button>
                            </Card>
                            : projects.length === 0
                                ? <Typography variant="body2" color="text.secondary">No project matches.</Typography>
                                : <Grid container spacing={2}>
                                    {projects.map(p => {
                                        const apps = p.apps ?? [];
                                        return <Grid key={p.id} size={{xs: 12, sm: 6, md: 4}}>
                                            <Card variant="outlined" sx={{height: '100%', position: 'relative'}}>
                                                <ButtonBase component={RouterLink} to={`/project/${p.identifier}`}
                                                            sx={{display: 'block', textAlign: 'left', p: 2.5, height: '100%', width: '100%'}}>
                                                    <Stack direction="row" spacing={1} alignItems="center" sx={{pr: 4}}>
                                                        <Typography component="h2" variant="h6" noWrap>{p.name}</Typography>
                                                        {p.identifier === last && <Chip size="small" variant="outlined" label="last opened"/>}
                                                    </Stack>
                                                    <Typography variant="body2" color="text.secondary" noWrap sx={{fontFamily: 'monospace', mb: 2}}>{p.identifier}</Typography>
                                                    {apps.length === 0
                                                        ? <Typography variant="body2" color="text.secondary">No apps yet</Typography>
                                                        : <Stack spacing={0.75}>
                                                            {apps.slice(0, 4).map(a => <Stack key={a.id} direction="row" spacing={1} alignItems="center">
                                                                <PlatformTypeIcon platformType={a.platform} fontSize="small" color="action"/>
                                                                <Typography variant="body2" noWrap>{a.name}</Typography>
                                                            </Stack>)}
                                                            {apps.length > 4 && <Typography variant="caption" color="text.secondary">and {apps.length - 4} more</Typography>}
                                                        </Stack>}
                                                </ButtonBase>
                                                <Tooltip title="Settings">
                                                    <IconButton size="small" aria-label={`Settings of ${p.name}`} sx={{position: 'absolute', top: 12, right: 12}}
                                                                component={RouterLink} to={`/project/${p.identifier}/settings`}>
                                                        <SettingsRounded fontSize="small"/>
                                                    </IconButton>
                                                </Tooltip>
                                            </Card>
                                        </Grid>;
                                    })}
                                </Grid>}
            </Box>

            <NewProjectDialog open={creating} onClose={() => setCreating(false)} onCreated={project => {
                dispatch({type: HubStateActions.ProjectsUpdate, payload: [...(state.projects ?? []), project]});
                setCreating(false);
                // a new project needs an app first
                navigate(`/project/${project.identifier}/settings?tab=apps`);
            }}/>
        </>
    );
};

export default ProjectsIndexPage;
