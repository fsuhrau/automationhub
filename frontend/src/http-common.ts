import axios from 'axios';

export default axios.create({
    baseURL: `${process.env.NODE_ENV == 'development' ? 'http://localhost:8002' : '' }/api`,
    withCredentials: true,
    headers: {
        'Content-type': 'application/json',
    },
});
