import React, { FC, MouseEvent, useEffect } from 'react';
import { createStyles, Theme, WithStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Paper from '@material-ui/core/Paper';
import { Button, Typography, withStyles } from '@material-ui/core';
import IAppData from '../types/app';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        root: {
            margin: 'auto',
        },
        paper: {
            width: 200,
            height: 230,
            overflow: 'auto',
        },
        button: {
            margin: theme.spacing(0.5, 0),
        },
        input: {
            display: 'none',
        },
    });

interface AppSelectionProps extends WithStyles<typeof styles> {
    onSelectionChanged: (app: IAppData) => void;
    apps: IAppData[];
    upload: boolean;
}

const AppSelection: FC<AppSelectionProps> = (props) => {
    const { classes, onSelectionChanged, apps, upload } = props;

    const [app, setApp] = React.useState<IAppData>();

    const handleSelect = (id: number | null | undefined, e: MouseEvent<HTMLDivElement>): void => {
        e.preventDefault();
        for (let i = 0; i < apps.length; i++) {
            if (apps[ i ].ID === id) {
                setApp(apps[ i ]);
                break;
            }
        }
    };

    useEffect(() => {
        if (app !== undefined) {
            onSelectionChanged(app);
        }
    }, [app, onSelectionChanged]);

    return (
        <Grid
            container={ true }
            spacing={ 2 }
            justifyContent="center"
            alignItems="center"
            direction={ 'column' }
            className={ classes.root }
        >
            { upload && (
                <Grid item={ true }>
                    <Typography variant={ 'subtitle1' }>Upload new App</Typography>
                    <input
                        accept="*.apk,*.ipa"
                        className={ classes.input }
                        id="app-upload"
                        multiple={ true }
                        type="file"
                    />
                    <label htmlFor="app-upload">
                        <Button variant="contained"
                            color="primary"
                            component="span">
                            Upload
                        </Button>
                    </label>
                </Grid>
            ) }
            <Grid item={ true }>
                <Typography variant={ 'subtitle1' }>Available Apps</Typography>
                <Paper className={ classes.paper }>
                    <List dense={ true } component="div" role="list">
                        { apps.map((a) => {
                            <ListItem key={ 'appid_' + a.ID } role="listitem" button={ true }
                                onClick={ (e) => handleSelect(a.ID, e) }>
                                <ListItemText key={ 'appt_' + a.ID } id={ a.Name }
                                    primary={ `${ a.Name }(${ a.Version })` }>{ a.Name }({ a.Version })</ListItemText>
                            </ListItem>;
                        }) }
                    </List>
                </Paper>
            </Grid>
        </Grid>
    );
};

export default withStyles(styles)(AppSelection);
