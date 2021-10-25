import { FC } from 'react';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import {
    Button,
    Card, CardActions, CardContent,
    CardMedia,
    ImageList,
    ImageListItem,
    ImageListItemBar, Typography,
} from '@material-ui/core';
import IProtocolEntryData from '../types/protocol.entry';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        root: {
            display: 'flex',
            flexWrap: 'wrap',
            justifyContent: 'space-around',
            overflow: 'hidden',
            backgroundColor: theme.palette.background.paper,
        },
        imageList: {
            flexWrap: 'nowrap',
            // Promote the list into his own layer on Chrome. This cost memory but helps keeping high FPS.
            transform: 'translateZ(0)',
        },
        title: {
            color: theme.palette.primary.light,
        },
        titleBar: {
            background:
                'linear-gradient(to top, rgba(0,0,0,0.7) 0%, rgba(0,0,0,0.3) 70%, rgba(0,0,0,0) 100%)',
        },
    });

interface ProtocolEntriesProps extends WithStyles<typeof styles> {
    entries: IProtocolEntryData[]
}

const ProtocolScreensComponent: FC<ProtocolEntriesProps> = (props) => {
    const { entries, classes } = props;
    return (
        <div className={classes.root}>
            <ImageList className={classes.imageList} rowHeight={500} cols={1}>
                {entries.map((item) => (
                    <ImageListItem key={'image_' + item.ID}>
                        <Card >
                            <CardMedia
                                component="img"
                                height="350"
                                image={`/api/data/${item.Data}`}
                                alt={item.Data}
                            />
                        </Card>
                    </ImageListItem>
                ))}
            </ImageList>
        </div>
    );
};

export default withStyles(styles)(ProtocolScreensComponent);
