-- Test main
function main(attrs, fn)
    print(attrs["uuid"])
    return { foo = "bar" }, nil
end