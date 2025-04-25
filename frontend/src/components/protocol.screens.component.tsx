import React from 'react';

import { Card, CardMedia, ImageList, ImageListItem } from '@mui/material';
import IProtocolEntryData from '../types/protocol.entry';

interface ProtocolEntriesProps {
    entries: IProtocolEntryData[]
}

const ProtocolScreensComponent: React.FC<ProtocolEntriesProps> = (props) => {
    const { entries } = props;
    return (
        <ImageList rowHeight={ 500 } cols={ 1 }>
            { entries.map((item) => (
                <ImageListItem key={ 'image_' + item.id }>
                    <Card>
                        <CardMedia
                            component="img"
                            height="350"
                            image={ `/api/data/${ item.data }` }
                            alt={ item.data }
                        />
                    </Card>
                </ImageListItem>
            )) }
        </ImageList>
    );
};

export default ProtocolScreensComponent;
