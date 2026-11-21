import React, {useState} from 'react';
import {Box, FormControl, InputLabel, MenuItem, Select, Stack, ToggleButton, ToggleButtonGroup} from '@mui/material';
import {Section} from '../../../components/common';
import IProject from '../../../project/project';
import {ICreatedPersonalToken} from '../../../types/profile.token';
import {HubLocation, keychainCommand, mcpCommand, proxyCommand, TOKEN_PLACEHOLDER} from '../profileModel';
import GuideStep from './GuideStep';

type Setup = 'container' | 'direct';

interface StepsProps {
    where: HubLocation,
    project: string,
    token: ICreatedPersonalToken | null,
}

const ContainerSteps: React.FC<StepsProps> = ({where, project}) => <>
    <GuideStep title="Keep the token on the host, in the Keychain" code={keychainCommand(where.host)}>
        Containers must not hold the token. The command asks for it.
    </GuideStep>
    <GuideStep title="Run the MCP proxy on the host" code={proxyCommand(where.origin, project)}>
        It adds the token to Claude Code's requests; the container connects to the proxy instead of the hub.
    </GuideStep>
</>;

const DirectStep: React.FC<StepsProps> = ({where, project, token}) =>
    <GuideStep title="Add the hub to Claude Code" code={mcpCommand(where.origin, project, token?.token ?? TOKEN_PLACEHOLDER)}>
        {token ? 'Your new token is filled in.' : `Put your token in place of ${TOKEN_PLACEHOLDER}.`}
    </GuideStep>;

const ProjectSelect: React.FC<{ projects: IProject[], value: string, onChange: (identifier: string) => void }> = ({projects, value, onChange}) =>
    <FormControl size="small" sx={{minWidth: 200}}>
        <InputLabel id="guide-project">Project</InputLabel>
        <Select labelId="guide-project" label="Project" value={value} onChange={e => onChange(e.target.value)}>
            {projects.map(p => <MenuItem key={p.identifier} value={p.identifier}>{p.name}</MenuItem>)}
        </Select>
    </FormControl>;

export interface ConnectGuideProps {
    projects: IProject[],
    token: ICreatedPersonalToken | null,
    where: HubLocation,
}

// ConnectGuide walks through connecting Claude Code to the hub with a
// personal token; with a token just created, its value is filled in
const ConnectGuide: React.FC<ConnectGuideProps> = ({projects, token, where}) => {
    const [setup, setSetup] = useState<Setup>('container');
    const [picked, setPicked] = useState('');
    const project = picked || projects[0]?.identifier || '';
    const steps = {where, project: project || '<project>', token};

    return <Section title="Connect Claude Code" description="Let Claude Code run tests and control devices through the hub's MCP server, as you.">
        <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} alignItems={{sm: 'center'}} sx={{mb: 2}}>
            <ToggleButtonGroup size="small" exclusive={true} value={setup} onChange={(_, v) => v && setSetup(v)} aria-label="Where Claude Code runs">
                <ToggleButton value="container">In a dev container</ToggleButton>
                <ToggleButton value="direct">Directly on this computer</ToggleButton>
            </ToggleButtonGroup>
            {projects.length > 0 && <ProjectSelect projects={projects} value={project} onChange={setPicked}/>}
        </Stack>
        <Box component="ol" sx={{m: 0, p: 0, listStyle: 'none', counterReset: 'step'}}>
            <GuideStep title={`Create a personal token above${token ? ` (done: ${token.name})` : ''}`}>
                It's shown once; copy it when it appears.
            </GuideStep>
            {setup === 'container' ? <ContainerSteps {...steps}/> : <DirectStep {...steps}/>}
        </Box>
    </Section>;
};

export default ConnectGuide;
