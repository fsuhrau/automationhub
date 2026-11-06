import React from "react";
import {Tooltip} from "@mui/material";
import {Person, SmartToy} from "@mui/icons-material";

export interface DeviceControllerInfo {
    // who controls the device, for display
    label: string,
    // true when an MCP agent holds the device rather than a person
    agent: boolean,
}

// describeController turns a lease owner into something readable. Owners are
// a user name, "token:<name>" for API tokens, or "mcp:<caller>:<session>" for
// MCP agents (the caller is a user name or "token:<name>").
export const describeController = (owner: string): DeviceControllerInfo => {
    if (owner.startsWith('mcp:')) {
        let caller = owner.substring('mcp:'.length);
        const sessionSep = caller.lastIndexOf(':');
        if (sessionSep > 0) {
            caller = caller.substring(0, sessionSep);
        }
        return {label: `agent of ${describeController(caller).label}`, agent: true};
    }
    if (owner.startsWith('token:')) {
        return {label: `token ${owner.substring('token:'.length)}`, agent: false};
    }
    return {label: owner, agent: false};
};

// DeviceController shows who controls the device live, if anyone, as a
// colored icon with the details in its tooltip.
const DeviceController: React.FC<{ owner?: string }> = ({owner}) => {
    if (!owner) {
        return null;
    }
    const {label, agent} = describeController(owner);
    const Icon = agent ? SmartToy : Person;
    return (
        <Tooltip title={<><b>Controlled by {label}</b>{owner !== label && <><br/>Lease owner: {owner}</>}</>}
                 arrow placement="left">
            <Icon fontSize="small" color="warning" sx={{cursor: 'help'}} aria-label={`Controlled by ${label}`}/>
        </Tooltip>
    );
};

export default DeviceController;
