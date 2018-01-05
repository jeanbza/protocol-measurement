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

        const interval = setInterval(() => fetch(new Request('/runs'))
            .then(resp => resp.json())
            .then(runs => this.setState({runs, loading: false}))
            .catch(err => console.error(err)), 200)

        this.state = {
            loading: true,
            runs: [],
            interval: interval,
            selectedAmount: '100',
        }
    }

    componentWillUnmount() {
        clearInterval(this.state.interval)
    }

    handleSubmit() {
        console.log('Sending', this.state.selectedAmount)
        fetch(new Request('/runs', {method: 'POST', body: `{"numMessages":${this.state.selectedAmount}}`}))
            .then(resp => resp.json())
            .then(json => {
                console.log(json)
                this.props.onRunChange(json.id)
            })
            .catch(err => console.error(err))
    }

    render() {
        const {runs, loading} = this.state
        const {selectedRun, onRunChange} = this.props

        const sidebarItems = runs.map((run, index) =>
            <SidebarItem key={index}
                         title={run}
                         active={run === selectedRun}
                         onClick={() => onRunChange(run)}/>)

        let content = <div className="sidebar--loading">
            Loading runs...
        </div>

        if (!loading) {
            if (sidebarItems.length > 0) {
                content = <div className="sidebar--items">
                    {sidebarItems}
                </div>
            } else {
                content = <div className="sidebar--loading">
                    No runs have been created!
                </div>
            }
        }

        return <div className="sidebar">
            <h3 className="sidebar--title">
                Run a new run
            </h3>
            <div className="new-run">
                <label>Messages</label>
                <select name="numMessages" onChange={e => this.setState({selectedAmount: e.target.value})}>
                    <option value="100">100</option>
                    <option value="1000">1,000</option>
                    <option value="10000">10,000</option>
                    <option value="100000">100,000</option>
                    <option value="1000000">1,000,000</option>
                </select>
                <button onClick={this.handleSubmit.bind(this)}>Run new run</button>
            </div>
            <h3 className="sidebar--title">
                Select a run
            </h3>
            {content}
        </div>
    }
}