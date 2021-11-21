import axios from 'axios';

export default axios.create({
    baseURL: `${process.env.NODE_ENV == 'development' ? 'http://localhost:8002' : '' }/api`,
    headers: {
        'Content-type': 'application/json',
    },
});
