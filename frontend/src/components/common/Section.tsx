import React from 'react';
import {Box, Card, CardContent, Stack, Typography} from '@mui/material';

export interface SectionProps {
    title: string,
    description?: React.ReactNode,
    action?: React.ReactNode,
    // stretch to the height of the grid cell
    fullHeight?: boolean,
    children: React.ReactNode,
}

// Section is the titled card the pages are made of
export const Section: React.FC<SectionProps> = ({title, description, action, fullHeight, children}) =>
    <Card variant="outlined" sx={fullHeight ? {height: '100%'} : undefined}>
        <CardContent>
            <Stack direction="row" alignItems="flex-start" spacing={1} useFlexGap sx={{mb: 1.5}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                    {description && <Typography variant="body2" color="text.secondary">{description}</Typography>}
                </Box>
                {action && <Box sx={{ml: 'auto', flexShrink: 0}}>{action}</Box>}
            </Stack>
            {children}
        </CardContent>
    </Card>;
