import React from "react";
import {Typography} from "@mui/material";
import Grid from "@mui/material/Grid2";
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
            {title && <Typography component="h2" variant="h6" sx={{mb: 2, mt: 2}}>
                {title}
            </Typography>}
            {titleElement && <Box sx={{mb: 2, mt: 2}}>
                {titleElement}
            </Box>}
            <Grid container spacing={2} columns={1}>
                <Grid size={{xs: 12, lg: 12}}>
                    {children}
                </Grid>
            </Grid>
        </>
    )
};