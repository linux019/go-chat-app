import moment from 'moment';
import React from 'react';
import PropTypes from 'prop-types';

class ChatDialogue extends React.Component {
    state = {
        messages: [],
        text: '',
    };

    static propTypes = {
        activeChannel: PropTypes.string.isRequired,
        setCallback: PropTypes.func.isRequired,
        sendUserMessage: PropTypes.func.isRequired,
    };

    componentDidMount() {
        this.props.setCallback(this.storeMessages);
    }

    storeMessages = messages => this.setState({messages});

    componentWillUnmount() {
        this.props.setCallback(null);
    }

    onTextChange = e => this.setState({text: e.target.value});

    onSubmit = e => {
        e.preventDefault();
        e.stopPropagation();
        this.props.sendUserMessage(this.state.text);
        this.setState({text: ''});
    };

    render() {
        const {messages, text} = this.state;
        return (
            <div className="dialogue">
                <div className={'chat'}>
                    <div className={'messages'}>
                        {
                            messages.length > 0
                                ? messages.map(message =>
                                    <ChatMessage key={`${message.sender}-${message.time}`} {...message}/>)
                                : <p className={'center'}>No messages yet</p>
                        }
                    </div>
                    <div className={'text-input'}>
                        <textarea value={text}
                                  onChange={this.onTextChange}
                        />
                        <button onClick={this.onSubmit} disabled={!text}>Submit</button>
                    </div>
                </div>
            </div>
        )
    }
}

function dateFormat(unixtime) {
    const msgDate = new Date();
    msgDate.setTime(unixtime * 1000);
    const isSame = msgDate.getFullYear() === new Date().getFullYear();
    const momentDate = moment.unix(unixtime);
    const format = {
        sameDay: 'h:mm A',
        lastDay: '[Yesterday], h:mm A',
        lastWeek: 'MMM D, h:mm A',
        sameElse: 'MMM D, h:mm A'
    };
    return isSame ? momentDate.calendar(null, format) : momentDate.format('MMM D, YYYY h:mm A');
}

const ChatMessage = ({time, sender, message}) => (
    <div className='message'>
        <div className='header'>
            <span className='sender'>{sender}</span>
            <span className='time'>{dateFormat(time)}</span>
        </div>
        <div className='text'>{message}</div>
    </div>
);

export default ChatDialogue