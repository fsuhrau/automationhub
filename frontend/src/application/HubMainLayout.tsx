import React from 'react';
import {SSEProvider} from 'react-hooks-sse';
import {createHubEventSource} from "./sseSource";
import {ProjectProvider} from "../hooks/ProjectProvider";
import HubStateUpdates from './HubStateUpdates';
import LayoutShell from "../app/LayoutShell";

const hubEventSource = createHubEventSource('/api/sse');

// HubMainLayout is the frame of every project page. HubStateUpdates loads the
// project list and keeps the project's state up to date; the project provider
// waits for that list.
const HubMainLayout: React.FC = () => <SSEProvider source={hubEventSource}>
    <HubStateUpdates>
        <ProjectProvider>
            <LayoutShell/>
        </ProjectProvider>
    </HubStateUpdates>
</SSEProvider>;

export default HubMainLayout;
