import React, { FC, useEffect } from 'react';
import { makeStyles, Theme, createStyles, WithStyles, withStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Checkbox from '@material-ui/core/Checkbox';
import Button from '@material-ui/core/Button';
import Paper from '@material-ui/core/Paper';
import IDeviceData from "../types/device";
import DeviceDataService from "../services/device.service";
import { Typography } from "@material-ui/core";

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
    });

interface DeviceSelectionProps extends WithStyles<typeof styles> {
    onSelectionChanged: (devices: IDeviceData[]) => void;
    selectedDevices: IDeviceData[];
}
function not(a: IDeviceData[], b: IDeviceData[]) {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: IDeviceData[], b: IDeviceData[]) {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const DeviceSelection: FC<DeviceSelectionProps> = (props) => {
    const { classes, onSelectionChanged, selectedDevices } = props;

    const [checked, setChecked] = React.useState<IDeviceData[]>([]);

    const [left, setLeft] = React.useState<IDeviceData[]>([]);
    const [right, setRight] = React.useState<IDeviceData[]>([]);

    const leftChecked = intersection(checked, left);
    const rightChecked = intersection(checked, right);

    const handleToggle = (value: IDeviceData) => () => {
        const currentIndex = checked.indexOf(value);
        const newChecked = [...checked];

        if (currentIndex === -1) {
            newChecked.push(value);
        } else {
            newChecked.splice(currentIndex, 1);
        }

        setChecked(newChecked);
    };

    const handleAllRight = () => {
        setRight(right.concat(left));
        setLeft([]);
    };

    const handleCheckedRight = () => {
        setRight(right.concat(leftChecked));
        setLeft(not(left, leftChecked));
        setChecked(not(checked, leftChecked));
    };

    const handleCheckedLeft = () => {
        setLeft(left.concat(rightChecked));
        setRight(not(right, rightChecked));
        setChecked(not(checked, rightChecked));
    };

    const handleAllLeft = () => {
        setLeft(left.concat(right));
        setRight([]);
    };

    useEffect(() => {
        DeviceDataService.getAll().then(response => {
            setLeft(response.data);
        }).catch(e => {
        });
    }, []);

    useEffect(() => {
        onSelectionChanged(right);
    }, [right])

    const customList = (items: IDeviceData[]) => (
        <Paper className={classes.paper}>
            <List dense component="div" role="list">
                {items.map((value) => {
                    const labelId = `transfer-list-item-${value.ID}-label`;

                    return (
                        <ListItem key={value.ID} role="listitem" button onClick={handleToggle(value)}>
                            <ListItemIcon>
                                <Checkbox
                                    checked={checked.indexOf(value) !== -1}
                                    tabIndex={-1}
                                    disableRipple
                                    inputProps={{ 'aria-labelledby': labelId }}
                                />
                            </ListItemIcon>
                            <ListItemText id={labelId} primary={`${value.Name}`} />
                        </ListItem>
                    );
                })}
                <ListItem />
            </List>
        </Paper>
    );

    return (
        <Grid
            container
            spacing={2}
            justifyContent="center"
            alignItems="center"
            className={classes.root}
        >
            <Grid item>
                <Typography variant={"subtitle1"}>Available Devices</Typography>
                {customList(left)}
            </Grid>
            <Grid item>
                <Grid container direction="column" alignItems="center">
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleAllRight}
                        disabled={left.length === 0}
                        aria-label="move all right"
                    >
                        ≫
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleCheckedRight}
                        disabled={leftChecked.length === 0}
                        aria-label="move selected right"
                    >
                        &gt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleCheckedLeft}
                        disabled={rightChecked.length === 0}
                        aria-label="move selected left"
                    >
                        &lt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleAllLeft}
                        disabled={right.length === 0}
                        aria-label="move all left"
                    >
                        ≪
                    </Button>
                </Grid>
            </Grid>
            <Grid item>
                <Typography variant={"subtitle1"}>Selected Devices</Typography>
                {customList(right)}
            </Grid>
        </Grid>
    );
}

export default withStyles(styles)(DeviceSelection);
