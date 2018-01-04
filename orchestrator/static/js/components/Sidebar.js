import React from 'react'

class SidebarItem extends React.Component {
    render() {
        const {num, active, onClick} = this.props

        const className = active ? "sidebar--item sidebar--item__active" : "sidebar--item sidebar--item__inactive"

        return <div className={className} onClick={onClick}>
            Sidebar item {num}
        </div>
    }
}

export default class Sidebar extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            activeSet: undefined,
        }
    }

    render() {
        const {activeSet} = this.state
        const items = []

        for (let i = 0; i < 20; i++) {
            items.push(<SidebarItem key={i}
                                    num={i}
                                    active={i === activeSet}
                                    onClick={() => this.setState({activeSet: i})}/>)
        }

        return <div className="sidebar">
            <h3 className="sidebar--title">
                Select a set
            </h3>
            <div className="sidebar--items">
                {items}
            </div>
        </div>
    }
}