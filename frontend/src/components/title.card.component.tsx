import React from "react";
import {Typography} from "@mui/material";
import Grid from "@mui/material/Grid";
import {Box} from "@mui/system";

interface TitleCardProps {
    title?: string,
    titleElement?: React.ReactElement,
    // one line on what the section is for
    subtitle?: React.ReactNode,
    // the main actions of the section, shown on the right of the title
    actions?: React.ReactNode,
    children?: React.ReactNode,
}

export const TitleCard: React.FC<TitleCardProps> = (props: TitleCardProps) => {
    const {title, titleElement, subtitle, actions, children} = props;
    return (
        <>
            <Box sx={{mb: 2, mt: 2, display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap'}}>
                <Box sx={{minWidth: 0}}>
                    <Box sx={{display: 'flex', gap: 2, alignItems: 'center'}}>
                        {title && <Typography component="h2" variant="h6">
                            {title}
                        </Typography>}
                        {titleElement}
                    </Box>
                    {subtitle && <Typography variant="body2" color="text.secondary">{subtitle}</Typography>}
                </Box>
                {actions && <Box sx={{ml: 'auto', display: 'flex', gap: 1, alignItems: 'center'}}>{actions}</Box>}
            </Box>
            <Grid container spacing={2} columns={1}>
                <Grid size={{xs: 12, lg: 12}}>
                    {children}
                </Grid>
            </Grid>
        </>
    )
};
