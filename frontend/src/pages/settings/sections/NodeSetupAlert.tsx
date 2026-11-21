import React from 'react';
import {Alert, Typography} from '@mui/material';
import {Code} from '../../../components/common';
import {NodeWithToken} from '../../../types/settings.requests';
import {nodeConfig} from '../settingsModel';

export interface NodeSetupAlertProps {
    node: NodeWithToken,
    // the master's address as this browser reaches it; a node may need another
    master: string,
    onClose: () => void,
}

// NodeSetupAlert shows the config a node needs to connect; the token only now
const NodeSetupAlert: React.FC<NodeSetupAlertProps> = ({node, master, onClose}) =>
    <Alert severity="success" sx={{mb: 2, '& .MuiAlert-message': {flex: 1, minWidth: 0}}} onClose={onClose}>
        <Typography variant="body2" sx={{mb: 1}}>
            Put this into the <code>automationhub.yaml</code> of <b>{node.name}</b> and start it with <code>automationhub node</code>.
            The token is only shown now.
        </Typography>
        <Code label="Copy the config">{nodeConfig(node.identifier, node.token, master)}</Code>
        <Typography variant="caption" color="text.secondary">
            <code>master_url</code> is the address this browser uses; change it if the node reaches the master differently.
        </Typography>
    </Alert>;

export default NodeSetupAlert;
