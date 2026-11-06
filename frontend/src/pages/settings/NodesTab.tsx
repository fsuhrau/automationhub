import React, {useEffect, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Chip,
    IconButton,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {DeleteOutlineRounded, KeyRounded} from '@mui/icons-material';
import {createNode, deleteNode, getNodes, issueNodeToken, NodeWithToken} from '../../services/settings.service';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from '../../ErrorProvider';
import {INodeData} from '../../types/node';
import {Code, ConfirmDialog, CopyButton, messageOf, Section} from './parts';
import {nodeConfig} from './settingsModel';

// NodesTab manages the computers that connect devices to this hub: adding
// one gives the config it needs, once
const NodesTab: React.FC = () => {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError();
    const [name, setName] = useState('');
    const [tried, setTried] = useState(false);
    const [busy, setBusy] = useState(false);
    const [setup, setSetup] = useState<NodeWithToken | null>(null);
    const [deleting, setDeleting] = useState<INodeData | null>(null);
    const [reissuing, setReissuing] = useState<INodeData | null>(null);
    // the master's address as this browser reaches it; a node may need another
    const master = window.location.host;

    useEffect(() => {
        getNodes(projectIdentifier).then(list => dispatch({type: HubStateActions.NodesUpdate, payload: list}))
            .catch(e => setError(messageOf(e)));
    }, [projectIdentifier]);

    const create = () => {
        setTried(true);
        if (!name.trim()) return;
        setBusy(true);
        createNode(projectIdentifier, {name: name.trim()}).then(node => {
            dispatch({type: HubStateActions.NodeAdd, payload: node});
            setSetup(node);
            setName('');
            setTried(false);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const reissue = () => {
        if (!reissuing) return;
        setBusy(true);
        issueNodeToken(projectIdentifier, reissuing.id).then(node => {
            setSetup(node);
            setReissuing(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const remove = () => {
        if (!deleting) return;
        setBusy(true);
        deleteNode(projectIdentifier, deleting.id).then(() => {
            dispatch({type: HubStateActions.NodeDelete, payload: deleting.id});
            if (setup?.id === deleting.id) setSetup(null);
            setDeleting(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const nodes = state.nodes ?? [];

    return <Section title="Nodes" description="Computers that connect phones, simulators, editors and browsers to this hub. Nodes are shared by all projects.">
        <Box component="form" noValidate onSubmit={e => {
            e.preventDefault();
            create();
        }} sx={{mb: 2}}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'flex-start'}}>
                <TextField size="small" label="Name" placeholder="e.g. Mac mini QA" value={name} onChange={e => setName(e.target.value)}
                           error={tried && !name.trim()} helperText={tried && !name.trim() ? 'A name is needed.' : 'Shown in the device pool for its devices.'}
                           sx={{minWidth: 240}}/>
                <Button type="submit" variant="contained" disabled={busy}>Add node</Button>
            </Stack>
        </Box>

        {setup && <Alert severity="success" sx={{mb: 2, '& .MuiAlert-message': {flex: 1, minWidth: 0}}} onClose={() => setSetup(null)}>
            <Typography variant="body2" sx={{mb: 1}}>
                Put this into the <code>automationhub.yaml</code> of <b>{setup.name}</b> and start it with <code>automationhub node</code>.
                The token is only shown now.
            </Typography>
            <Code label="Copy the config">{nodeConfig(setup.identifier, setup.token, master)}</Code>
            <Typography variant="caption" color="text.secondary">
                <code>master_url</code> is the address this browser uses; change it if the node reaches the master differently.
            </Typography>
        </Alert>}

        {nodes.length === 0
            ? <Typography variant="body2" color="text.secondary">No nodes yet. Devices plugged into the master itself work without one.</Typography>
            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Nodes">
                <TableHead><TableRow>
                    <TableCell>Name</TableCell><TableCell>Status</TableCell><TableCell>Identifier</TableCell><TableCell/>
                </TableRow></TableHead>
                <TableBody>
                    {nodes.map(n => <TableRow key={n.id}>
                        <TableCell>{n.name}</TableCell>
                        <TableCell>
                            {n.status === 1
                                ? <Chip size="small" color="success" label="Connected"/>
                                : <Chip size="small" label="Offline"/>}
                        </TableCell>
                        <TableCell sx={{whiteSpace: 'nowrap'}}>
                            <Stack direction="row" alignItems="center" spacing={0.5}>
                                <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{n.identifier}</Typography>
                                <CopyButton text={n.identifier} label="Copy the identifier"/>
                            </Stack>
                        </TableCell>
                        <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                            <Tooltip title="New token, to set the node up again">
                                <IconButton size="small" aria-label={`New token for ${n.name}`} onClick={() => setReissuing(n)}><KeyRounded fontSize="small"/></IconButton>
                            </Tooltip>
                            <Tooltip title="Delete">
                                <IconButton size="small" aria-label={`Delete ${n.name}`} onClick={() => setDeleting(n)}><DeleteOutlineRounded fontSize="small"/></IconButton>
                            </Tooltip>
                        </TableCell>
                    </TableRow>)}
                </TableBody>
            </Table></Box>}

        <ConfirmDialog open={!!reissuing} title={`New token for ${reissuing?.name}?`} confirm="Issue new token" color="warning" busy={busy}
                       onConfirm={reissue} onClose={() => setReissuing(null)}>
            The node's current token stops working, so it disconnects until its config has the new one.
        </ConfirmDialog>
        <ConfirmDialog open={!!deleting} title={`Delete ${deleting?.name}?`} confirm="Delete node" busy={busy}
                       onConfirm={remove} onClose={() => setDeleting(null)}>
            Its token stops working, so it can't connect anymore. Its devices stay in the device pool but can't be used through it.
        </ConfirmDialog>
    </Section>;
};

export default NodesTab;
