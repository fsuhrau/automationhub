import React from 'react';
import {Typography} from '@mui/material';
import {ConfirmDialog, Section} from '../../components/common';
import NodeForm from './sections/NodeForm';
import NodeSetupAlert from './sections/NodeSetupAlert';
import NodesTable from './sections/NodesTable';
import {useConfirmAction} from './useConfirmAction';
import {useNodes} from './useNodes';

// NodesTab manages the computers that connect devices to this hub: adding
// one gives the config it needs, once
const NodesTab: React.FC = () => {
    const {nodes, setup, closeSetup, create, reissue, remove} = useNodes();
    const reissuing = useConfirmAction(reissue);
    const deleting = useConfirmAction(remove);

    return <Section title="Nodes" description="Computers that connect phones, simulators, editors and browsers to this hub. Nodes are shared by all projects.">
        <NodeForm busy={create.busy} onCreate={create.run}/>
        {setup && <NodeSetupAlert node={setup} master={window.location.host} onClose={closeSetup}/>}
        {nodes.length === 0
            ? <Typography variant="body2" color="text.secondary">No nodes yet. Devices plugged into the master itself work without one.</Typography>
            : <NodesTable nodes={nodes} onReissue={reissuing.ask} onDelete={deleting.ask}/>}
        <ConfirmDialog open={!!reissuing.target} title={`New token for ${reissuing.target?.name}?`} confirm="Issue new token" color="warning"
            busy={reissuing.busy} onConfirm={reissuing.confirm} onClose={reissuing.cancel}>
            The node's current token stops working, so it disconnects until its config has the new one.
        </ConfirmDialog>
        <ConfirmDialog open={!!deleting.target} title={`Delete ${deleting.target?.name}?`} confirm="Delete node"
            busy={deleting.busy} onConfirm={deleting.confirm} onClose={deleting.cancel}>
            Its token stops working, so it can't connect anymore. Its devices stay in the device pool but can't be used through it.
        </ConfirmDialog>
    </Section>;
};

export default NodesTab;
