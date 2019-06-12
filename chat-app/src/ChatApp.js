import React from 'react';
import PropTypes from 'prop-types';

import {ChatDialogue, Sidebar} from './items';

export const DataContext = React.createContext({});
const serverAddress = 'ws://localhost:4488/ws';

class ChatApp extends React.Component {
    state = {
        activeChannel: null,
        connected: false,
        channels: [],
    };

    static propTypes = {
        userName: PropTypes.string.isRequired
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

        socket.onmessage = event => {
            console.log('DATA', event.data);
            try {
                const data = JSON.parse(event.data);
                if (data) {
                    this.onServerData(data);
                }
            } catch (e) {

            }
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

    onServerData = data => {
        const newState = {};
        if (data.channels) {
            newState.channels = data.channels;
            if (!this.state.activeChannel) {
                newState.activeChannel = data.channels[0];
            }
        }

        if (Object.keys(newState).length) {
            this.setState(newState);
        }
    };

    componentWillUnmount() {
        this.socket.close();
        this.socket = null;
        if (this.timeoutID) {
            clearTimeout(this.timeoutID);
        }
    }

    setActiveChannel = activeChannel => this.setState({activeChannel});

    render() {
        const {userName} = this.props;
        const {channels, connected, activeChannel} = this.state;
        return (
            <DataContext.Provider
                value={{userName, connected, channels, activeChannel, setActiveChannel: this.setActiveChannel}}>
                <Sidebar/>
                <ChatDialogue/>
            </DataContext.Provider>
        )
    }
}

export default ChatApp