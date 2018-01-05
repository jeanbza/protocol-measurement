import React from 'react'
import moment from 'moment'

class SidebarItem extends React.PureComponent {
    render() {
        const {title, timeCreated, active, onClick} = this.props

        const className = active ? "sidebar--item sidebar--item__active" : "sidebar--item sidebar--item__inactive"

        return <div className={className} onClick={onClick}>
            <div>{title}</div>
            <div>{moment().diff(timeCreated, 'minutes')}m ago</div>
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

        const sidebarItems = runs
            .sort((a, b) => moment(b.createdAt).unix() - moment(a.createdAt).unix())
            .map((run, index) => <SidebarItem key={index}
                                              title={run.id}
                                              timeCreated={run.createdAt}
                                              active={run.id === selectedRun}
                                              onClick={() => onRunChange(run.id)}/>)

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