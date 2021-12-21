import { FC } from 'react';
import { Card, CardMedia, ImageList, ImageListItem } from '@mui/material';
import IProtocolEntryData from '../types/protocol.entry';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
        flexWrap: 'wrap',
        justifyContent: 'space-around',
        overflow: 'hidden',
    },
    imageList: {
        flexWrap: 'nowrap',
        // Promote the list into his own layer on Chrome. This cost memory but helps keeping high FPS.
        transform: 'translateZ(0)',
    },
    titleBar: {
        background:
            'linear-gradient(to top, rgba(0,0,0,0.7) 0%, rgba(0,0,0,0.3) 70%, rgba(0,0,0,0) 100%)',
    },
}));

interface ProtocolEntriesProps {
    entries: IProtocolEntryData[]
}

const ProtocolScreensComponent: FC<ProtocolEntriesProps> = (props) => {
    const classes = useStyles();
    const { entries } = props;
    return (
        <ImageList className={ classes.imageList } rowHeight={ 500 } cols={ 1 }>
            { entries.map((item) => (
                <ImageListItem key={ 'image_' + item.ID }>
                    <Card>
                        <CardMedia
                            component="img"
                            height="350"
                            image={ `/api/data/${ item.Data }` }
                            alt={ item.Data }
                        />
                    </Card>
                </ImageListItem>
            )) }
        </ImageList>
    );
};

export default ProtocolScreensComponent;
