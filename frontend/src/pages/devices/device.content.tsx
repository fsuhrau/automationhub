import { FC } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import DeviceTableComponent from '../../components/device-table.component';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 1200,
            margin: 'auto',
            overflow: 'hidden',
        },
        searchBar: {
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
        },
        searchInput: {
            fontSize: theme.typography.fontSize,
        },
        block: {
            display: 'block',
        },
        addUser: {
            marginRight: theme.spacing(1),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
    });

export type DevicesProps = WithStyles<typeof styles>;

const Devices: FC<DevicesProps> = props => {
    const { classes } = props;

    return (
        <Paper className={classes.paper}>
            <DeviceTableComponent/>
        </Paper>
    );
};

export default withStyles(styles)(Devices);
