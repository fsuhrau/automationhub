import React from "react";
import { Typography } from "@mui/material";
import Grid from "@mui/material/Grid2";

interface TitleCardProps {
    title: string,
    children?: React.ReactNode,
}

export const TitleCard: React.FC<TitleCardProps> = (props: TitleCardProps) => {
    const {children, title} = props;
    return (
        <>
            <Typography component="h2" variant="h6" sx={{ mb: 2 }}>
                { title }
            </Typography>
            <Grid container spacing={2} columns={1}>
                <Grid size={{ xs: 12, lg: 12 }}>
                    { children }
                </Grid>
            </Grid>
        </>
    )
};