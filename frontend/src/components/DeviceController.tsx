import React from "react";
import {Tooltip} from "@mui/material";
import {Person, SmartToy} from "@mui/icons-material";

export interface DeviceControllerInfo {
    // who controls the device, for display
    label: string,
    // true when an MCP agent holds the device rather than a person
    agent: boolean,
    // the hub user id the lease belongs to, when the owner carries one
    userId?: number,
}

const USER_SUFFIX = /^(.*) \(user:(\d+)\)$/;
const BARE_USER = /^user:(\d+)$/;

const parseUser = (owner: string): { label: string, userId?: number } => {
    const suffixed = USER_SUFFIX.exec(owner);
    if (suffixed) return {label: suffixed[1], userId: Number(suffixed[2])};
    const bare = BARE_USER.exec(owner);
    if (bare) return {label: owner, userId: Number(bare[1])};
    return {label: owner};
};

// the caller of an MCP owner "mcp:<caller>:<session>"
const mcpCaller = (owner: string): string => {
    const caller = owner.substring('mcp:'.length);
    const sessionSep = caller.lastIndexOf(':');
    return sessionSep > 0 ? caller.substring(0, sessionSep) : caller;
};

// describeController parses a lease owner once into something readable.
// Owners are "<name> (user:<id>)" or "user:<id>" for signed-in users,
// "token:<name>" for API tokens, or "mcp:<caller>:<session>" for MCP agents
// (the caller is a user or "token:<name>").
export const describeController = (owner: string): DeviceControllerInfo => {
    if (owner.startsWith('mcp:')) {
        const {label, userId} = describeController(mcpCaller(owner));
        return {label: `agent of ${label}`, agent: true, userId};
    }
    if (owner.startsWith('token:')) {
        return {label: `token ${owner.substring('token:'.length)}`, agent: false};
    }
    return {...parseUser(owner), agent: false};
};

// DeviceController shows who controls the device live, if anyone, as a
// colored icon with the details in its tooltip.
const DeviceController: React.FC<{ owner?: string }> = ({owner}) => {
    if (!owner) {
        return null;
    }
    const {label, agent, userId} = describeController(owner);
    const Icon = agent ? SmartToy : Person;
    return (
        <Tooltip title={<><b>Controlled by {label}</b>{owner !== label && (agent || userId === undefined) && <><br/>Lease owner: {owner}</>}</>}
            arrow={true} placement="left">
            <Icon fontSize="small" color="warning" sx={{cursor: 'help'}} aria-label={`Controlled by ${label}`}/>
        </Tooltip>
    );
};

export default DeviceController;
