import React from "react";
import {Typography} from "@mui/material";
import Grid from "@mui/material/Grid";
import {Box} from "@mui/system";

interface TitleCardProps {
    title?: string,
    titleElement?: React.ReactElement,
    children?: React.ReactNode,
}

export const TitleCard: React.FC<TitleCardProps> = (props: TitleCardProps) => {
    const {title, titleElement, children} = props;
    return (
        <>
            <Box sx={{mb: 2, mt: 2, display: 'flex', gap: 2}}>
                {title && <Typography component="h2" variant="h6">
                    {title}
                </Typography>}
                {titleElement}
            </Box>
            <Grid container spacing={2} columns={1}>
                <Grid size={{xs: 12, lg: 12}}>
                    {children}
                </Grid>
            </Grid>
        </>
    )
};