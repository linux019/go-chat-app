import React from 'react';
import PropTypes from 'prop-types';

import {ChatDialogue, Sidebar} from './items';

export const DataContext = React.createContext({});
const serverAddress = 'ws://localhost:4488/ws';

class ChatApp extends React.Component {
    state = {
        connected: false,
        channels: [],
    };

    componentDidMount() {
        this.openServerConnection();
    }

    openServerConnection = () => {
        const socket = new WebSocket(serverAddress);
        socket.onopen = this.onConnectionOpen;

        socket.onclose = event => {
            let reconnect = false;
            if (event.wasClean) {
                console.log('WS:DISCONNECTED');
            } else {
                reconnect = true;
                console.log('WS:DISCONNECTED (abort)');
            }
            console.log(`WS:DISCONNECTED (${event.code} ${event.reason})`);
            this.onConnectionClose(reconnect);
        };

        socket.onmessage = function (event) {
            console.log('DATA', event.data)
        };

        socket.onerror = function (error) {
            console.log('WS:ERROR', error);
        };
        this.socket = socket;
    };

    onConnectionOpen = () => {
        console.log('WS:OK');
        this.setState({connected: true});
        this.timeoutID = null;


        this.setName();
        // this.getChannels();
    };

    onConnectionClose = reconnect => {
        this.setState({connected: false});
        this.socket = null;
        if (reconnect) {
            this.timeoutID = setTimeout(this.openServerConnection, 10000);
        }
    };

    getChannels = () => this.sendCommand('GET_CHANNELS', null);
    setName = () => this.sendCommand('SET_USERNAME', this.props.userName);

    sendCommand = (command, data) => this.socket.send(JSON.stringify({data, command}));

    componentWillUnmount() {
        this.socket.close();
        this.socket = null;
        if (this.timeoutID) {
            clearTimeout(this.timeoutID);
        }
    }

    static propTypes = {
        userName: PropTypes.string.isRequired
    };

    render() {
        const {userName} = this.props;
        const {channels, connected} = this.state;
        return (
            <DataContext.Provider value={{userName, connected}}>
                <Sidebar/>
                <ChatDialogue/>
            </DataContext.Provider>
        )
    }
}

export default ChatApp