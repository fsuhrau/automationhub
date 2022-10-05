import React from "react";
import Grid from "@mui/material/Grid";
import { Typography } from "@mui/material";

interface TitleCardProps {
    title: string,
    children?: React.ReactNode,
}

export const TitleCard: React.FC<TitleCardProps> = (props: TitleCardProps) => {
    const {children, title} = props;
    return (
        <Grid container={ true } spacing={ 1 } sx={ {marginTop: 5} }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "overline" }>{ title }</Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                { children }
            </Grid>
        </Grid>
    )
};