import React from 'react'

class SidebarItem extends React.Component {
    render() {
        const {title, active, onClick} = this.props

        const className = active ? "sidebar--item sidebar--item__active" : "sidebar--item sidebar--item__inactive"

        return <div className={className} onClick={onClick}>
            {title}
            <div className="sidebar--item-border"/>
        </div>
    }
}

export default class Sidebar extends React.Component {
    constructor(props) {
        super(props)

        const interval = setInterval(() => fetch(new Request('/sets'))
            .then(resp => resp.json())
            .then(sets => this.setState({sets: sets, loading: false}))
            .catch(err => console.error(err)), 200)

        this.state = {
            loading: true,
            sets: [],
            interval: interval,
            selectedAmount: '1000',
        }
    }

    componentWillUnmount() {
        clearInterval(this.state.interval)
    }

    handleSubmit() {
        console.log('Sending', this.state.selectedAmount)
        fetch(new Request('/sets', {method: 'POST', body: `{"numMessages":${this.state.selectedAmount}}`}))
            .then(() => console.log('Done'))
            .catch(err => console.error(err))
    }

    render() {
        const {sets, loading} = this.state
        const {selectedSet, onSetChange} = this.props

        const sidebarItems = sets.map((set, index) =>
            <SidebarItem key={index}
                         title={set}
                         active={set === selectedSet}
                         onClick={() => onSetChange(set)}/>)

        let content = <div className="sidebar--loading">
            Loading sets...
        </div>

        if (!loading) {
            if (sidebarItems.length > 0) {
                content = <div className="sidebar--items">
                    {sidebarItems}
                </div>
            } else {
                content = <div className="sidebar--loading">
                    No sets have been created!
                </div>
            }
        }

        return <div className="sidebar">
            <h3 className="sidebar--title">
                Run a new set
            </h3>
            <div className="new-run">
                <label>Messages</label>
                <select name="numMessages" onChange={e => this.setState({selectedAmount: e.target.value})}>
                    <option value="1000">1,000</option>
                    <option value="10000">10,000</option>
                    <option value="100000">100,000</option>
                    <option value="1000000">1,000,000</option>
                </select>
                <button onClick={this.handleSubmit.bind(this)}>Run new set</button>
            </div>
            <h3 className="sidebar--title">
                Select a set
            </h3>
            {content}
        </div>
    }
}