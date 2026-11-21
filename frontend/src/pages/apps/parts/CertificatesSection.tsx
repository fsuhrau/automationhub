import React from 'react';
import {Alert, Box, Chip, Stack, Typography} from '@mui/material';
import {Section} from '../../../components/common';
import {Certificate} from '../../../types/bundle.info';
import {commonName} from '../bundleDetailModel';
import {certificateFacts, signingNote} from '../signingFactsModel';
import {Facts} from './Facts';

const CertificateCard: React.FC<{ certificate: Certificate, now: number }> = ({certificate, now}) =>
    <Box sx={{mb: 1.5, '&:last-child': {mb: 0}}}>
        <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 0.5}}>
            <Typography variant="subtitle2">{commonName(certificate.subject)}</Typography>
            {certificate.debug && <Chip size="small" color="warning" label="debug key"/>}
        </Stack>
        <Facts facts={certificateFacts(certificate)} now={now}/>
    </Box>;

interface CertificatesSectionProps {
    certificates: Certificate[],
    schemes?: string[],
    apple: boolean,
    now: number,
}

export const CertificatesSection: React.FC<CertificatesSectionProps> = ({certificates, schemes, apple, now}) =>
    <Section title="Signing" description={signingNote(apple, schemes)} fullHeight={true}>
        {certificates.length === 0 && <Alert severity="warning">The bundle is not signed; devices won't install it.</Alert>}
        {certificates.map(c => <CertificateCard key={c.sha256} certificate={c} now={now}/>)}
    </Section>;
