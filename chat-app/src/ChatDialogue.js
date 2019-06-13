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
        this.props.sendUserMessage({message: this.state.text});
    };

    render() {
        const {messages, text} = this.state;
        return (
            <div className="dialogue">
                <div className={'chat'}>
                    <div className={'messages'}>
                        {
                            messages.length > 0 ? '' : <p className={'center'}>No messages yet</p>
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

export default ChatDialogue