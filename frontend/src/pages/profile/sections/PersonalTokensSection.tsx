import React from 'react';
import {Alert, Typography} from '@mui/material';
import {Code, Section} from '../../../components/common';
import {ICreatedPersonalToken} from '../../../types/profile.token';
import PersonalTokenForm, {PersonalTokenFormProps} from './PersonalTokenForm';
import PersonalTokensTable, {PersonalTokensTableProps} from './PersonalTokensTable';

const Description: React.FC = () => <>
    A personal token acts as you in every project: for Claude Code (MCP) and scripts, sent in the <code>X-Auth-Token</code> header.
    Device locks show your name. The value is shown once, when you create it.
</>;

const CreatedToken: React.FC<{ token: ICreatedPersonalToken, onClose: () => void }> = ({token, onClose}) =>
    <Alert severity="success" onClose={onClose} sx={{mb: 2, '& .MuiAlert-message': {flex: 1, minWidth: 0}}}>
        <Typography variant="body2" sx={{mb: 1}}>
            <b>{token.name}</b> is ready. Copy it now; it won't be shown again. The guide below has it filled in.
        </Typography>
        <Code label="Copy the token">{token.token}</Code>
    </Alert>;

export interface PersonalTokensSectionProps extends Omit<PersonalTokensTableProps, 'tokens'>, PersonalTokenFormProps {
    // null until loaded
    tokens: PersonalTokensTableProps['tokens'] | null,
    created: ICreatedPersonalToken | null,
    onDismissCreated: () => void,
}

// PersonalTokensSection creates and revokes the tokens that act as the person
const PersonalTokensSection: React.FC<PersonalTokensSectionProps> = ({tokens, created, onDismissCreated, busy, onCreate, ...table}) =>
    <Section title="Personal access tokens" description={<Description/>}>
        <PersonalTokenForm busy={busy} onCreate={onCreate}/>
        {created && <CreatedToken token={created} onClose={onDismissCreated}/>}
        {tokens !== null && (tokens.length === 0
            ? <Typography variant="body2" color="text.secondary">No personal tokens yet.</Typography>
            : <PersonalTokensTable tokens={tokens} {...table}/>)}
    </Section>;

export default PersonalTokensSection;
