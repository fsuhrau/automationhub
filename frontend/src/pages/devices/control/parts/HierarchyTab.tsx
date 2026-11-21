import React from 'react';
import {Alert, Box, Button, Stack, ToggleButton, Tooltip, Typography} from '@mui/material';
import {RefreshRounded} from '@mui/icons-material';
import {RichTreeView} from '@mui/x-tree-view/RichTreeView';
import {SceneControl} from '../useSceneGraph';
import {ControlSession} from '../useControlSession';
import ElementDetails from './ElementDetails';

interface HierarchyTabProps {
    session: ControlSession,
    control: SceneControl,
    onAppTab: () => void,
}

const SceneTree: React.FC<{ control: SceneControl }> = ({control}) => {
    const {scene, selected, expanded} = control;
    if (!scene) return null;
    return (
        <Box sx={{maxHeight: '50vh', overflow: 'auto'}}>
            <RichTreeView items={scene.items} expandedItems={expanded}
                selectedItems={selected ? scene.idOf.get(selected) ?? null : null}
                onExpandedItemsChange={(_, ids) => control.setExpanded(ids)}
                onSelectedItemsChange={(_, id) => control.select(id ? scene.byId.get(id as string) ?? null : null)}/>
        </Box>
    );
};

// the scene graph of the running app, for finding elements and their XPath
const HierarchyTab: React.FC<HierarchyTabProps> = ({session, control, onAppTab}) => (
    <Box>
        <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
            The scene graph of the running app, for finding elements and their XPath.
            Needs the app connected to the hub (start it from the App tab).
        </Typography>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}}>
            <Button variant="outlined" startIcon={<RefreshRounded/>} onClick={control.load} disabled={session.busy}>
                {control.scene ? 'Reload' : 'Load scene graph'}
            </Button>
            <Tooltip title="Click an element on the screen to find it in the tree; the click doesn't reach the device">
                <span><ToggleButton size="small" value="pick" selected={control.picking} disabled={!control.scene}
                    onChange={control.togglePicking}>Pick on screen</ToggleButton></span>
            </Tooltip>
        </Stack>
        {control.error && <Alert severity="warning" sx={{mb: 1.5}}
            action={<Button color="inherit" size="small" onClick={onAppTab}>App tab</Button>}>
            {control.error}
        </Alert>}
        {control.selected && <ElementDetails element={control.selected} session={session}/>}
        <SceneTree control={control}/>
    </Box>
);

export default HierarchyTab;
