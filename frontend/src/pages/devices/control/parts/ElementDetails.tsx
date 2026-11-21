import React from 'react';
import {Box, Button, Paper, Stack, Typography} from '@mui/material';
import {copyText} from '../../../../lib/clipboard';
import {nodeLabel, xpathFor} from '../sceneModel';
import {ControlSession} from '../useControlSession';

interface ElementDetailsProps {
    element: Element,
    session: ControlSession,
}

const mono = {fontFamily: 'monospace', wordBreak: 'break-all'};

const Attributes: React.FC<{ element: Element }> = ({element}) => (
    <Box sx={{mt: 1, display: 'grid', gridTemplateColumns: 'max-content 1fr', columnGap: 2, rowGap: 0.25}}>
        {Array.from(element.attributes).map(a => <React.Fragment key={a.name}>
            <Typography variant="caption" color="text.secondary">{a.name}</Typography>
            <Typography variant="caption" sx={mono}>{a.value}</Typography>
        </React.Fragment>)}
    </Box>
);

// the selected element of the scene graph: tap it, copy its XPath, see its attributes
const ElementDetails: React.FC<ElementDetailsProps> = ({element, session}) => {
    const {service, run, report, lease: {token}} = session;
    const id = element.getAttribute('ID');
    const xpath = xpathFor(element);
    const tap = (): void => {
        if (token && id) run(() => service.touchElement(token, {elementId: id}));
    };
    const copy = async (): Promise<void> => {
        if (!await copyText(xpath)) report('Could not copy to the clipboard');
    };
    return (
        <Paper variant="outlined" sx={{p: 1.5, mb: 1.5}}>
            <Stack direction="row" spacing={1} alignItems="center" useFlexGap={true}>
                <Typography variant="subtitle2" sx={{wordBreak: 'break-word'}}>{nodeLabel(element)}</Typography>
                <Button size="small" variant="contained" sx={{ml: 'auto'}} disabled={!token || !id} onClick={tap}>Tap it</Button>
            </Stack>
            <Stack direction="row" spacing={1} alignItems="center" sx={{mt: 1}}>
                <Typography variant="body2" sx={{...mono, flex: 1}}>{xpath}</Typography>
                <Button size="small" onClick={copy}>Copy XPath</Button>
            </Stack>
            <Attributes element={element}/>
        </Paper>
    );
};

export default ElementDetails;
