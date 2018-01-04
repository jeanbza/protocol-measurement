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

        this.state = {
            loading: true,
            activeSet: undefined,
            sets: [],
        }

        fetch(new Request('/sets'))
            .then(resp => resp.json())
            .then(sets => this.setState({sets: sets, loading: false}))
            .catch(err => console.error(err))
    }

    render() {
        const {sets, activeSet, loading} = this.state

        const sidebarItems = sets.map((set, index) =>
            <SidebarItem key={index}
                         title={set}
                         active={set === activeSet}
                         onClick={() => this.setState({activeSet: set})}/>)

        let content = <div className="sidebar--loading">
            Loading sets...
        </div>

        if (!loading) {
            content = <div className="sidebar--items">
                {sidebarItems}
            </div>
        }

        return <div className="sidebar">
            <h3 className="sidebar--title">
                Select a set
            </h3>
            {content}
        </div>
    }
}